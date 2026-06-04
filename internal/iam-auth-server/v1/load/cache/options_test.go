package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOptions_DefaultValues(t *testing.T) {
	opts := &Options{}
	assert.Equal(t, int64(0), opts.NumCounters)
	assert.Equal(t, int64(0), opts.MaxCost)
	assert.Equal(t, int64(0), opts.BufferItems)
}

func TestOptions_WithValues(t *testing.T) {
	opts := &Options{
		NumCounters: 1e7,     // 10M
		MaxCost:     1 << 30, // 1GB
		BufferItems: 64,
	}
	assert.Equal(t, int64(1e7), opts.NumCounters)
	assert.Equal(t, int64(1<<30), opts.MaxCost)
	assert.Equal(t, int64(64), opts.BufferItems)
}
