package service

import (
	"github.com/casbin/casbin/v2"
	"go.uber.org/atomic"
	"sync"
	"time"
)

type EnforcerCache struct {
	mtx   sync.RWMutex
	cache map[string]*cachedEnforcer
	TTL   time.Duration
	pool  *sync.Pool
}

type cachedEnforcer struct {
	e        *casbin.Enforcer
	LastTime atomic.Time
}

func NewEnforcerCache(TTL time.Duration, pool *sync.Pool) *EnforcerCache {
	cache := &EnforcerCache{
		cache: make(map[string]*cachedEnforcer),
		TTL:   TTL,
		mtx:   sync.RWMutex{},
		pool:  pool,
	}

	go cache.clearLoop()

	return cache
}

func (c *EnforcerCache) Get(secretID string) (*casbin.Enforcer, bool) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	e, ok := c.cache[secretID]
	e.LastTime.Store(time.Now())
	return e.e, ok
}

func (c *EnforcerCache) Set(secretID string, enforcer *casbin.Enforcer) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	e := cachedEnforcer{e: enforcer}
	e.LastTime.Store(time.Now())
	c.cache[secretID] = &e
}

func (c *EnforcerCache) clearLoop() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			var toDelete []string
			c.mtx.RLock()
			for key, value := range c.cache {
				if value != nil {
					if time.Since(value.LastTime.Load()) > c.TTL {
						toDelete = append(toDelete, key)
					}
				}
			}
			c.mtx.RUnlock()

			c.mtx.Lock()
			for _, k := range toDelete {
				c.pool.Put(c.cache[k].e)
				delete(c.cache, k)
			}
			c.mtx.Unlock()
		}
	}
}
