package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/storage"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"strconv"
	"sync"
	"time"
)

const RedisKeyPrefix string = "analytics-"
const defaultRedisAddress string = "127.0.0.1:6379"

var (
	redisClusterSingleton *RedisClusterStorage
	mtx                   sync.Mutex
	syncOnce              sync.Once
)

type RedisClusterStorage struct {
	db         redis.UniversalClient
	keyPrefix  string
	Config     genericoptions.RedisOptions
	stopChan   <-chan struct{}
	expireTime time.Duration
}

var _ storage.AnalyticsStore = &RedisClusterStorage{}

func NewRedisClusterStorage() *RedisClusterStorage {
	return &RedisClusterStorage{}
}

func Init(config interface{}) (*RedisClusterStorage, error) {
	syncOnce.Do(func() {
		redisClusterSingleton = &RedisClusterStorage{}
		err := mapstructure.Decode(config, &redisClusterSingleton.Config)
		if err != nil {
			log.Errorf("Error decoding config: %v", err)
			return
		}
		redisClusterSingleton.db = newRedisClusterPool(true, redisClusterSingleton.Config)
		redisClusterSingleton.keyPrefix = RedisKeyPrefix
	})
	mtx.Lock()
	defer mtx.Unlock()
	if redisClusterSingleton == nil {
		return nil, errors.New("redisClusterSingleton init error")
	}
	return redisClusterSingleton, nil
}

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

func newRedisClusterPool(forceReconnect bool, conf genericoptions.RedisOptions) redis.UniversalClient {
	mtx.Lock()
	defer mtx.Unlock()
	if !forceReconnect && redisClusterSingleton != nil {
		log.Debug("Redis pool already INITIALIZED")
		return redisClusterSingleton.db
	}

	if forceReconnect && redisClusterSingleton != nil {
		_ = redisClusterSingleton.db.Close()
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
	redisClusterSingleton.db = client

	return client
}
func (r *RedisClusterStorage) Connect() error {
	if r.db == nil {
		log.Debug("Connecting to Redis")
		r.db = newRedisClusterPool(false, r.Config)
		return nil
	}

	log.Debug("Redis already INITIALIZED")
	r.db = redisClusterSingleton.db
	return nil
}

func (r *RedisClusterStorage) SetKeyPrefix(keyPrefix string) {
	r.keyPrefix = keyPrefix
}

func (r *RedisClusterStorage) fixKey(rawKey string) string {
	return r.keyPrefix + rawKey
}

func (r *RedisClusterStorage) EnsureConnection() error {
	if r.db != nil {
		// 验证连接是否真的可用
		ctx, cancel := context.WithTimeout(context.Background(), r.Config.ReadTimeout)
		defer cancel()
		if _, err := r.db.Ping(ctx).Result(); err == nil {
			return nil
		}
		// 连接无效，重置
		r.db = nil
	}

	// 使用指数退避，最多重试10次
	backoff := time.Second
	maxBackoff := time.Second * 30
	maxRetries := 30

	for i := 0; i < maxRetries; i++ {
		select {
		case <-r.stopChan:
			return nil
		default:
		}

		// 尝试连接
		if err := r.Connect(); err == nil {
			// 连接成功，验证一下
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			if _, pingErr := r.db.Ping(ctx).Result(); pingErr == nil {
				log.Info("Redis connection established successfully")
				cancel()
				return nil
			} else {
				// Ping 失败，继续重试
				cancel()
				log.Warnf("Ping failed after Connect: %v", pingErr)
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

		case <-r.stopChan:
			return nil
		}
	}

	return nil
}

func (r *RedisClusterStorage) WithStopChan(stopChan <-chan struct{}) storage.AnalyticsStore {
	r.stopChan = stopChan
	return r
}

func (r *RedisClusterStorage) WithExpireTime(expireTime time.Duration) storage.AnalyticsStore {
	r.expireTime = expireTime
	return r
}

func (r *RedisClusterStorage) GetExpire(key string) (time.Duration, error) {
	fixedKey := r.fixKey(key)
	if err := r.EnsureConnection(); err != nil {
		log.Errorf("Redis connection failed: %v", err)
		return 0, err
	}

	value, err := r.db.TTL(context.Background(), fixedKey).Result()
	if err != nil {
		log.Errorf("Redis TTL failed: %v", err)
		return 0, err
	}
	return value, nil
}

func (r *RedisClusterStorage) SetExpire(key string, duration time.Duration) error {
	fixedKey := r.fixKey(key)
	if err := r.EnsureConnection(); err != nil {
		log.Errorf("Redis connection failed: %v", err)
		return err
	}
	err := r.db.Expire(context.Background(), fixedKey, duration).Err()
	if err != nil {
		log.Errorf("Redis Expire failed: %v", err)
		return err
	}
	return nil
}

func (r *RedisClusterStorage) AppendToSetPipelined(key string, values [][]byte) error {
	if len(values) == 0 {
		return nil
	}

	fixedKey := r.fixKey(key)
	if err := r.EnsureConnection(); err != nil {
		return err
	}
	client := r.db
	pipeline := client.Pipeline()
	ctx, cancel := context.WithTimeout(context.Background(), r.Config.WriteTimeout)
	defer cancel()
	for _, rawValue := range values {
		pipeline.RPush(ctx, fixedKey, rawValue)
	}
	if _, err := pipeline.Exec(ctx); err != nil {
		log.Errorf("Error trying to append to set keys:%s,%v", fixedKey, err.Error())
		return err
	}

	if r.expireTime > 0 {
		exp, _ := r.GetExpire(key)
		// 看一下有没有设置，没有的话补上
		if exp == -1 {
			err := r.SetExpire(key, r.expireTime)
			if err != nil {
				log.Errorf("Error setting expire time: %v", err)
			}
		}
	}

	return nil
}

func GetRedisClusterSingleton() *RedisClusterStorage {
	mtx.Lock()
	defer mtx.Unlock()
	return redisClusterSingleton
}

func (r *RedisClusterStorage) RegisterPubSubHandler(channel string, callback func(v interface{})) error {
	err := r.EnsureConnection()
	if err != nil {
		return err
	}
	pubsub := r.db.Subscribe(context.Background(), channel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(context.Background()); err != nil {
		log.Errorf("Redis PubSub Receive failed: %v", err)
		return err
	}

	for msg := range pubsub.Channel() {
		callback(msg)
	}
	return nil
}
