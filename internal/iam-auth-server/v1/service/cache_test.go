package service

import (
	"sync"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/stretchr/testify/require"
)

func newTestEnforcerPool() *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			m, err := casbinmodel.NewModelFromString(CurrenCasbinModelString)
			if err != nil {
				panic(err)
			}
			e, err := casbin.NewEnforcer(m)
			if err != nil {
				panic(err)
			}
			return e
		},
	}
}

func TestEnforcerCacheClear(t *testing.T) {
	pool := newTestEnforcerPool()
	c := NewEnforcerCache(time.Second*5, pool)

	e1, _ := pool.Get().(*casbin.Enforcer)
	c.Set("secret-1", e1)
	e2, _ := pool.Get().(*casbin.Enforcer)
	c.Set("secret-2", e2)

	_, ok := c.Get("secret-1")
	require.True(t, ok)
	_, ok = c.Get("secret-2")
	require.True(t, ok)

	c.Clear()

	_, ok = c.Get("secret-1")
	require.False(t, ok, "Clear 后缓存应被清空")
	_, ok = c.Get("secret-2")
	require.False(t, ok, "Clear 后缓存应被清空")
}
