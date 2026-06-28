package pumps

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"os"
	"strconv"
	"strings"
	"time"
)

// Define unit constant.
const (
	_   = iota // ignore zero iota
	KiB = 1 << (10 * iota)
	MiB
	GiB
	TiB
)

type MongoPump struct {
	CommonPump
	client *mongo.Client
	config *MongoConf
	dbName string
}

var _ PumpInterface = (*MongoPump)(nil)

type MongoType string

const (
	StandardMongo = MongoType("Standard")
	AWSMongo      = MongoType("AWS")
)

type BaseMongoConf struct {
	URL                   string    `mapstructure:"url"`
	UseSSL                bool      `mapstructure:"use_ssl"`
	SSLInsecureSkipVerify bool      `mapstructure:"ssl_skip_verify"`
	SSLAllowInvalidHosts  bool      `mapstructure:"ssl_allow_invalid_hosts"`
	SSLCAFile             string    `mapstructure:"ssl_ca_file"`
	SSLKEMKeyFile         string    `mapstructure:"ssl_kem_key_file"`
	DBType                MongoType `mapstructure:"db_type"`
}

type MongoConf struct {
	BaseMongoConf
	CollectionName            string `mapstructure:"collection_name"`
	MaxInsertBatchSizeBytes   int    `mapstructure:"max_insert_batch_size_bytes"`
	MaxDocumentSizeBytes      int    `mapstructure:"max_document_size_bytes"`
	CollectionCapMaxSizeBytes int    `mapstructure:"collection_cap_max_size_bytes"`
	CollectionCapEnabled      bool   `mapstructure:"collection_cap_enabled"`
}

func loadCertificateAndKeyFromFile(path string) (*tls.Certificate, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("fail to read file %s, err: %s", path, err)
		return nil, err
	}

	var certificate tls.Certificate
	for {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			certificate.Certificate = append(certificate.Certificate, block.Bytes)
		} else {
			certificate.PrivateKey, err = parsePrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
		}
		raw = rest
	}

	if len(certificate.Certificate) == 0 {
		return nil, fmt.Errorf("fail to load certificate from %s", path)
	} else if certificate.PrivateKey == nil {
		return nil, fmt.Errorf("fail to load certificate from %s", path)
	}
	return &certificate, nil
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse private key")
}

// New 用于被工厂调用，创建同类型的Pump实例
func (m *MongoPump) New() PumpInterface {
	return &MongoPump{}
}

// GetName returns the mongo pump name.
func (m *MongoPump) GetName() string {
	return "MongoDB Pump"
}

func (m *MongoPump) Init(config interface{}) error {
	m.config = &MongoConf{}
	err := mapstructure.Decode(config, &m.config)
	if err != nil {
		log.Fatalf("fail to decode config: %s", err)
		return err
	}
	err = mapstructure.Decode(config, &m.config.BaseMongoConf)
	if err != nil {
		log.Fatalf("fail to decode config: %s", err)
		return err
	}
	log.Infof("Initing MongoPump, URL:%s, collection:%s ", m.config.URL, m.config.CollectionName)

	if m.config.MaxInsertBatchSizeBytes == 0 {
		log.Info("-- No max batch size set, defaulting to 10MB")
		m.config.MaxInsertBatchSizeBytes = 10 * MiB
	}

	if m.config.MaxDocumentSizeBytes == 0 {
		log.Info("-- No max document size set, defaulting to 10MB")
		m.config.MaxDocumentSizeBytes = 10 * MiB
	}

	err = m.connect()
	if err != nil {
		log.Fatalf("fail to connect: %s", err)
		return err
	}

	err = m.capCollection()
	if err != nil {
		log.Warnf("fail to cap collection: %s", err)
		return err
	}
	log.Debugf("MongoDB DB CS: %s", m.config.URL)
	log.Debugf("MongoDB Col: %s", m.config.CollectionName)
	return nil
}

func (m *MongoPump) connect() error {
	clientOpts, err := m.mongoDialInfo(m.config.BaseMongoConf)
	if err != nil {
		log.Errorf("fail to connect to mongodb: %s", err)
		return err
	}

	// Extract database name from URI
	m.dbName = extractDatabaseName(m.config.URL)

	m.client, err = mongo.Connect(clientOpts)
	if err != nil {
		log.Errorf("fail to connect to mongodb: %s, but retrying once", err)
		time.Sleep(5 * time.Second)
		m.client, err = mongo.Connect(clientOpts)
		if err != nil {
			log.Errorf("fail to connect to mongodb: %s", err)
			return err
		}
	}

	if m.config.DBType == "" {
		m.config.DBType = MongoType(queryMongoType(m.client))
	}

	return nil
}

func (m *MongoPump) mongoDialInfo(conf BaseMongoConf) (*options.ClientOptions, error) {
	clientOpts := options.Client().ApplyURI(m.config.URL)
	clientOpts.SetTimeout(30 * time.Second)

	if conf.UseSSL {
		tlsConfig := &tls.Config{}
		if conf.SSLInsecureSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		if conf.SSLCAFile != "" {
			caCert, err := os.ReadFile(conf.SSLCAFile)
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("fail to read CA file: %s", conf.SSLCAFile))
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
		if conf.SSLAllowInvalidHosts {
			tlsConfig.InsecureSkipVerify = true
			tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
				certs := make([]*x509.Certificate, len(rawCerts))
				for i, asn1Data := range rawCerts {
					cert, err := x509.ParseCertificate(asn1Data)
					if err != nil {
						return err
					}
					certs[i] = cert
				}

				opts := x509.VerifyOptions{
					Roots:         tlsConfig.RootCAs,
					CurrentTime:   time.Now(),
					DNSName:       "", // <- skip hostname verification
					Intermediates: x509.NewCertPool(),
				}

				for i, cert := range certs {
					if i == 0 {
						continue
					}
					opts.Intermediates.AddCert(cert)
				}
				_, err := certs[0].Verify(opts)

				return err
			}
		}

		if conf.SSLKEMKeyFile != "" {
			cert, err := loadCertificateAndKeyFromFile(conf.SSLKEMKeyFile)
			if err != nil {
				log.Fatalf("fail to load keypair: %s", err)
				return nil, err
			}

			tlsConfig.Certificates = []tls.Certificate{*cert}
		}

		clientOpts.SetTLSConfig(tlsConfig)
	}

	return clientOpts, nil
}

// extractDatabaseName extracts the database name from a MongoDB connection URI.
// Falls back to the authSource query parameter if no database is in the path.
func extractDatabaseName(uri string) string {
	prefix := "mongodb://"
	if strings.HasPrefix(uri, "mongodb+srv://") {
		prefix = "mongodb+srv://"
	}

	rest := uri[len(prefix):]

	// Remove user:password@ part
	if atIndex := strings.LastIndex(rest, "@"); atIndex != -1 {
		rest = rest[atIndex+1:]
	}

	// Remove query parameters
	if qIndex := strings.Index(rest, "?"); qIndex != -1 {
		rest = rest[:qIndex]
	}

	// Remove options (semicolon separated)
	if sIndex := strings.Index(rest, ";"); sIndex != -1 {
		rest = rest[:sIndex]
	}

	// Now we should have host[:port][/dbname]
	if slashIndex := strings.Index(rest, "/"); slashIndex != -1 {
		dbName := rest[slashIndex+1:]
		if dbName != "" {
			return dbName
		}
	}

	// Fallback: try to extract from authSource query parameter
	if qIndex := strings.Index(uri, "?"); qIndex != -1 {
		for _, param := range strings.Split(uri[qIndex+1:], "&") {
			kv := strings.SplitN(param, "=", 2)
			if len(kv) == 2 && kv[0] == "authSource" {
				return kv[1]
			}
		}
	}

	return ""
}

func queryMongoType(client *mongo.Client) MongoType {
	// Querying for the features which 100% not supported by AWS DocumentDB
	var result struct {
		Code int `bson:"code"`
	}

	_ = client.Database("admin").RunCommand(context.Background(), bson.D{{Key: "features", Value: 1}}).Decode(&result)

	if result.Code == 303 {
		return AWSMongo
	}

	return StandardMongo
}

func (m *MongoPump) capCollection() error {
	if !m.config.CollectionCapEnabled {
		return nil
	}

	exists, err := m.collectionExists(m.config.CollectionName)
	if err != nil {
		log.Errorf("Unable to determine if collection (%s) exists. Not capping collection: %v", m.config.CollectionName, err)
		return err
	}
	if exists {
		log.Warnf("Collection (%s) already exists. Capping could result in data loss. Ignoring", m.config.CollectionName)

		return nil
	}

	if strconv.IntSize < 64 {
		log.Warn("Pump running < 64bit architecture. Not capping collection as max size would be 2gb")

		return errors.New("Pump running < 64bit architecture")
	}

	if m.config.CollectionCapMaxSizeBytes <= 0 {
		m.config.CollectionCapMaxSizeBytes = 5 * GiB
		log.Infof("-- No max collection size set for %s, defaulting to %d", m.config.CollectionName, m.config.CollectionCapMaxSizeBytes)
	}

	err = m.client.Database(m.dbName).CreateCollection(context.Background(),
		m.config.CollectionName,
		options.CreateCollection().SetCapped(true).SetSizeInBytes(int64(m.config.CollectionCapMaxSizeBytes)),
	)
	if err != nil {
		log.Errorf("fail to create collection: %s", err)
		return err
	}

	log.Infof("Capped collection (%s) created.", m.config.CollectionName)
	return nil
}

func (m *MongoPump) collectionExists(name string) (bool, error) {
	colNames, err := m.client.Database(m.dbName).ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		log.Fatalf("fail to get collection names: %s", err)
		return false, err
	}

	for _, colName := range colNames {
		if colName == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *MongoPump) WriteData(ctx context.Context, keyValues []interface{}) error {
	collectionName := m.config.CollectionName
	if collectionName == "" {
		log.Fatalf("No collection name!")
		return errors.New("No collection name")
	}

	log.Debugf("Writing data to collection: %s", collectionName)

	for m.client == nil {
		log.Debugf("Waiting for session to come up...")
		err := m.connect()
		if err != nil {
			time.Sleep(5 * time.Second)
		}
	}

	for _, dataSet := range m.accumulateSet(keyValues) {
		// 并行写入数据
		go func(dataSet []interface{}) {
			collection := m.client.Database(m.dbName).Collection(collectionName)
			_, err := collection.InsertMany(context.Background(), dataSet)
			if err != nil {
				log.Errorf("Error inserting data to collection: %s", err)
				if strings.Contains(strings.ToLower(err.Error()), "closed explicitly") {
					log.Warn("--> Detected connection failure!")
				}
			}
		}(dataSet)
	}

	return nil
}

func (m *MongoPump) accumulateSet(data []interface{}) [][]interface{} {
	accumulateToltal := 0
	resArray := make([][]interface{}, 0)
	thisResSet := make([]interface{}, 0)

	for i, item := range data {
		thisItem := item.(analytics.AnalyticsRecode)

		// Add 1 KB for metadata as average
		sizeBytes := len(thisItem.Reason) + 1024
		log.Debugf("%s size if %d", thisItem.SecretID, sizeBytes)
		if sizeBytes > m.config.MaxDocumentSizeBytes {
			log.Warn("Document too large, not writing raw request and raw response!")
			thisItem.Reason = ""
		}
		if accumulateToltal+sizeBytes < m.config.MaxInsertBatchSizeBytes {
			accumulateToltal += sizeBytes
		} else {
			if len(thisResSet) > 0 {
				resArray = append(resArray, thisResSet)
			}
			log.Debug("Created new chunk entry")
			thisResSet = make([]interface{}, 0)
			accumulateToltal = sizeBytes
		}

		thisResSet = append(thisResSet, item)

		if len(data)-1 == i {
			log.Debug("Appending last entry")
			resArray = append(resArray, thisResSet)
		}
	}
	return resArray
}
