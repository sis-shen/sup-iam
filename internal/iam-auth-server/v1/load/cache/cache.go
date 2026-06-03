package cache

import (
	"context"
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
	cachedmodel "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/rpc"
	"sync"
	"sync/atomic"
)

// 轻读重写用atomic
type Cache struct {
	cli         rpc.RpcClient
	secrets     atomic.Value
	policies    atomic.Value
	cacheConfig *ristretto.Config
}

var (
	ErrCacheMiss      = errors.New("cache miss")
	ErrSecretNotFound = errors.New("secret not found")
	ErrPolicyNotFound = errors.New("policy not found")
)

var (
	onceCache sync.Once
	cacheIns  *Cache
)

func InitSingleton(cli rpc.RpcClient, opts *Options) (*Cache, error) {
	onceCache.Do(func() {
		config := &ristretto.Config{
			NumCounters: opts.NumCounters,
			MaxCost:     opts.MaxCost,
			BufferItems: opts.BufferItems,
			Cost:        nil,
		}

		secretCache, err := ristretto.NewCache(config)
		if err != nil {
			return
		}
		policyCache, err := ristretto.NewCache(config)
		if err != nil {
			return
		}
		cacheIns = &Cache{
			cli:         cli,
			cacheConfig: config,
		}
		cacheIns.secrets.Store(secretCache)
		cacheIns.policies.Store(policyCache)
	})

	if cacheIns == nil {
		return nil, ErrCacheMiss
	}
	return cacheIns, nil
}

func (c *Cache) GetSecretByAK(accessKey string) (*cachedmodel.CachedSecret, error) {
	value, ok := c.secrets.Load().(*ristretto.Cache).Get(accessKey)
	if !ok {
		return nil, ErrSecretNotFound
	}
	return value.(*cachedmodel.CachedSecret), nil
}
func (c *Cache) GetPolicyListBySecretID(secretID string) ([]*cachedmodel.CachedPolicy, error) {
	value, ok := c.policies.Load().(*ristretto.Cache).Get(secretID)
	if !ok {
		return nil, ErrPolicyNotFound
	}
	return value.([]*cachedmodel.CachedPolicy), nil
}

func (c *Cache) ReloadSecrets() error {
	secrets, err := c.cli.GetAllSecrets(context.Background())
	if err != nil {
		return err
	}

	secretCache, err := ristretto.NewCache(c.cacheConfig)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		secretCache.Set(secret.AccessKey, secret, 1)
	}

	secretCache.Wait()
	c.secrets.Store(secretCache)
	//旧资源交给gc回收
	return nil
}

func (c *Cache) ReloadPolicies() error {

	policyGroups, err := c.cli.GetAllPolicies(context.Background())
	if err != nil {
		return err
	}

	policyCache, err := ristretto.NewCache(c.cacheConfig)
	if err != nil {
		return err
	}
	for _, policyGroup := range policyGroups {
		if len(policyGroup) == 0 {
			continue
		}
		policyCache.Set(policyGroup[0].SecretID, policyGroup, 1)
	}
	policyCache.Wait()

	c.policies.Store(policyCache)
	return nil
}

func (c *Cache) Reload() error {
	if err := c.ReloadSecrets(); err != nil {
		return err
	}
	if err := c.ReloadPolicies(); err != nil {
		return err
	}
	return nil
}
