package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"strconv"
	"sync"
	"time"
)

const defaultRedisAddress string = "127.0.0.1:6379"

var (
	redisClusterSingleton *RedisClusterStorage
	mtx                   sync.Mutex
	syncOnce              sync.Once
)

type RedisClusterStorage struct {
	db       redis.UniversalClient
	Config   genericoptions.RedisOptions
	stopChan <-chan struct{}
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
	mtx.Lock()
	defer mtx.Unlock()
	if redisClusterSingleton.db == nil {
		log.Debug("Connecting to Redis")
		r.db = newRedisClusterPool(false, r.Config)
		return nil
	}

	log.Debug("Redis already INITIALIZED")
	r.db = redisClusterSingleton.db
	return nil
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
		_ = r.db.Close()
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
		r.db = newRedisClusterPool(true, r.Config)
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

		case <-r.stopChan:
			return nil
		}
	}

	return nil
}

func (r *RedisClusterStorage) WithStopChan(stopChan <-chan struct{}) *RedisClusterStorage {
	r.stopChan = stopChan
	return r
}

func GetRedisClusterSingleton() *RedisClusterStorage {
	mtx.Lock()
	defer mtx.Unlock()
	return redisClusterSingleton
}

func (r *RedisClusterStorage) DB() redis.UniversalClient {
	return r.db
}
