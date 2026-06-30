package redis

import (
	"context"
	"crypto/tls"
	"github.com/go-viper/mapstructure/v2"
	"github.com/redis/go-redis/v9"
	storage "github.com/sis-shen/sup-iam/internal/iam-pump/store"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"strconv"
	"sync"
	"time"
)

const (
	RedisKeyPrefix      string = "analytics-"
	defaultRedisAddress string = "127.0.0.1:6379"
)

var (
	redisClusterSingleton redis.UniversalClient
	mtx                   sync.Mutex
)

// RedisClusterStorageManager is a storage manager that uses the redis database.
type RedisClusterStorageManager struct {
	db        redis.UniversalClient
	keyPrefix string
	Config    genericoptions.RedisOptions
	shutDown  chan bool
}

var _ storage.AnalyticStoreInterface = (*RedisClusterStorageManager)(nil)

func getAddrs(o *genericoptions.RedisOptions) (addrs []string) {
	if len(o.Addrs) > 0 {
		addrs = o.Addrs
	}

	if len(o.Addrs) == 0 && o.Port != 0 {
		addr := o.Host + ":" + strconv.FormatInt(int64(o.Port), 10)
		addrs = append(addrs, addr)
	}

	if len(o.Addrs) > 0 && o.Port == 0 {
		addrs = append(addrs, defaultRedisAddress)
	}

	return addrs
}

func NewRedisClusterPool(forceReconnect bool, conf genericoptions.RedisOptions) redis.UniversalClient {
	mtx.Lock()
	defer mtx.Unlock()
	if !forceReconnect && redisClusterSingleton != nil {
		log.Debug("Redis pool already INITIALIZED")
		return redisClusterSingleton
	}

	if forceReconnect && redisClusterSingleton != nil {
		_ = redisClusterSingleton.Close()
	}
	log.Debug("Creating new Redis Pool")

	var tlsConfig *tls.Config
	if conf.UseSSL {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: conf.SSLInsecureSkipVerify,
		}
	}

	var client redis.UniversalClient
	opts := &redis.UniversalOptions{
		Addrs:    getAddrs(&conf),
		Username: conf.Username,
		Password: conf.Password,
		DB:       conf.DB,

		// 连接池配置
		PoolSize:        conf.PoolSize,        // 连接池大小
		MaxActiveConns:  conf.MaxActiveConns,  // 最大活跃连接数
		MinIdleConns:    conf.MinIdleConns,    // 最小空闲连接数
		MaxIdleConns:    conf.MaxIdleConns,    // 最大空闲连接数
		ConnMaxIdleTime: conf.ConnMaxIdleTime, // 连接最大空闲时间
		ConnMaxLifetime: conf.ConnMaxLifetime, // 连接最大生命周期

		// 超时配置
		DialTimeout:  conf.DialTimeout,  // 连接超时
		ReadTimeout:  conf.ReadTimeout,  // 读超时
		WriteTimeout: conf.WriteTimeout, // 写超时
		PoolTimeout:  conf.PoolTimeout,  // 连接池获取连接超时

		//TLS设置
		TLSConfig: tlsConfig,
	}

	if conf.MasterName != "" {
		client = redis.NewFailoverClient(opts.Failover())
	} else if conf.EnableCluster {
		client = redis.NewClusterClient(opts.Cluster())
	} else {
		client = redis.NewClient(opts.Simple())
	}
	redisClusterSingleton = client

	return client
}

func (r *RedisClusterStorageManager) GetName() string {
	return "redis"
}

func (r *RedisClusterStorageManager) Init(conf interface{}) error {
	r.Config = genericoptions.RedisOptions{}
	err := mapstructure.Decode(conf, &r.Config)
	if err != nil {
		log.Errorf("Error decoding redis config: %v", err)
		return err
	}

	r.keyPrefix = RedisKeyPrefix
	return nil
}

func (r *RedisClusterStorageManager) Connect() error {
	if redisClusterSingleton == nil {
		log.Debug("Connecting to Redis")
		r.db = NewRedisClusterPool(false, r.Config)
		return nil
	}

	log.Debug("Redis already INITIALIZED")
	r.db = redisClusterSingleton
	return nil
}

func (r *RedisClusterStorageManager) SetKeyPrefix(keyPrefix string) {
	r.keyPrefix = keyPrefix
}

func (r *RedisClusterStorageManager) fixKey(rawKey string) string {
	return r.keyPrefix + rawKey
}

func (r *RedisClusterStorageManager) ensureConnection() error {
	if r.db != nil {
		// 验证连接是否真的可用
		ctx, cancel := context.WithTimeout(context.Background(), r.Config.ReadTimeout)
		defer cancel()
		if _, err := r.db.Ping(ctx).Result(); err == nil {
			return nil
		}
		// 连接无效，重置
		_ = r.db.Close()
		r.db = nil
	}

	// 使用指数退避，最多重试10次
	backoff := time.Second
	maxBackoff := time.Second * 30
	maxRetries := 30

	for i := 0; i < maxRetries; i++ {
		select {
		case <-r.shutDown:
			return nil
		default:
		}

		// 尝试连接
		r.db = NewRedisClusterPool(true, r.Config)
		if r.db != nil {

			if _, pingErr := r.db.Ping(context.Background()).Result(); pingErr == nil {
				log.Info("Redis connection established successfully")
				return nil
			} else {
				// Ping 失败，继续重试
				log.Warnf("Ping failed after Connect: %v", pingErr)
				_ = r.db.Close()
				r.db = nil
			}

		}
		log.Warnf("Redis atempting to Connect failed, try time: %d", i)

		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

		case <-r.shutDown:
			return nil
		}
	}

	return nil
}

func (r *RedisClusterStorageManager) WithShutDownChan(c chan bool) storage.AnalyticStoreInterface {
	r.shutDown = c
	return r
}

func (r *RedisClusterStorageManager) GetAndDeleteSet(keyName string) []interface{} {
	log.Debugf("Getting raw key set: %s", keyName)
	err := r.ensureConnection()
	if err != nil {
		return nil
	}

	fixedKey := r.fixKey(keyName)
	log.Debugf("Getting raw key: %s, fixed key:%s", keyName, fixedKey)
	var lrange *redis.StringSliceCmd
	ctx, cancel := context.WithTimeout(context.Background(), r.Config.ReadTimeout)
	defer cancel()
	_, err = r.db.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		lrange = pipe.LRange(ctx, fixedKey, 0, -1)
		pipe.Del(ctx, fixedKey)
		return nil
	})

	if err != nil {
		log.Errorf("Error getting key %s: %v", fixedKey, err)
		return nil
	}

	vals := lrange.Val()
	result := make([]interface{}, len(vals))
	for i, v := range vals {
		result[i] = v
	}
	log.Debugf("Unpacked vals: %d", len(result))
	return result
}
