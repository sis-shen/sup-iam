package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnalyticsKeyName(t *testing.T) {
	assert.Equal(t, "iam-system-analytics", AnalyticsKeyName)
}

func TestAnalyticStoreInterface(t *testing.T) {
	// Verify the interface exists by checking it can be implemented
	var _ AnalyticStoreInterface = (*mockStore)(nil)
}

// mockStore implements AnalyticStoreInterface for compile-time check
type mockStore struct{}

func (m *mockStore) Init(config interface{}) error                       { return nil }
func (m *mockStore) GetName() string                                     { return "mock" }
func (m *mockStore) Connect() error                                      { return nil }
func (m *mockStore) GetAndDeleteSet(string) []interface{}                { return nil }
func (m *mockStore) SetKeyPrefix(string)                                 {}
func (m *mockStore) WithShutDownChan(c chan bool) AnalyticStoreInterface { return m }

func TestAnalyticStoreInterface_WithShutDownChan(t *testing.T) {
	m := &mockStore{}
	c := make(chan bool, 1)
	result := m.WithShutDownChan(c)
	assert.Equal(t, m, result)
}
