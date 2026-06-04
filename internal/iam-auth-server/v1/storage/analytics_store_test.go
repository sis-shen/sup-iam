package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyticsKeyName(t *testing.T) {
	assert.Equal(t, "iam-system-analytics", AnalyticsKeyName)
}

func TestAnalyticsStoreInterface_Check(t *testing.T) {
	var _ AnalyticsStore = (*mockStore)(nil)
}

// mockStore implements AnalyticsStore for compile-time interface check
type mockStore struct{}

func (m *mockStore) Connect() error { return nil }

func (m *mockStore) AppendToSetPipelined(s string, b [][]byte) error { return nil }

func (m *mockStore) SetExpire(s string, d time.Duration) error { return nil }

func (m *mockStore) GetExpire(s string) (time.Duration, error) { return 0, nil }

func (m *mockStore) SetKeyPrefix(s string) {}

func (m *mockStore) WithStopChan(stopChan <-chan struct{}) AnalyticsStore { return m }

func (m *mockStore) WithExpireTime(d time.Duration) AnalyticsStore { return m }

func TestMockStore_InterfaceMethods(t *testing.T) {
	m := &mockStore{}

	assert.NoError(t, m.Connect())
	assert.NoError(t, m.AppendToSetPipelined("key", nil))
	assert.NoError(t, m.SetExpire("key", time.Hour))
	dur, err := m.GetExpire("key")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), dur)

	m.SetKeyPrefix("prefix")

	stopChan := make(chan struct{})
	result := m.WithStopChan(stopChan)
	assert.Equal(t, m, result)

	result2 := m.WithExpireTime(time.Hour)
	assert.Equal(t, m, result2)
}
