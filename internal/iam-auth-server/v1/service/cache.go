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
	if !ok {
		return nil, false
	}
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

// Clear 清空全部缓存，用于本地缓存收到更新信号并完成重载后保持策略一致性。
// 注意：与 clearLoop 不同，这里不把 enforcer 归还对象池、也不修改其内容——
// 重载信号无法判断缓存中的 enforcer 是否正被并发请求执行 Enforce，
// 直接丢弃引用交由 GC 回收可避免数据竞争；对象池中仅存放已 ClearPolicy 的空白 enforcer，无需清理。
func (c *EnforcerCache) Clear() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.cache = make(map[string]*cachedEnforcer)
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
				e := c.cache[k].e
				e.ClearPolicy()
				c.pool.Put(c.cache[k].e)
				delete(c.cache, k)
			}
			c.mtx.Unlock()
		}
	}
}
