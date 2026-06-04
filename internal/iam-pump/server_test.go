package iampump

import (
	"context"
	"testing"
	"time"

	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"github.com/sis-shen/sup-iam/internal/iam-pump/pumps"
	"github.com/stretchr/testify/assert"
)

// mockPump implements pumps.PumpInterface for testing filterData
type mockPump struct {
	name            string
	filters         *analytics.AnalyticFilters
	timeout         time.Duration
	omitDetail      bool
	writeDataCalled bool
	writtenKeys     []interface{}
}

func (m *mockPump) GetName() string          { return m.name }
func (m *mockPump) New() pumps.PumpInterface { return &mockPump{} }
func (m *mockPump) Init(interface{}) error   { return nil }
func (m *mockPump) WriteData(ctx context.Context, keys []interface{}) error {
	m.writeDataCalled = true
	m.writtenKeys = keys
	return nil
}
func (m *mockPump) GetFilters() *analytics.AnalyticFilters  { return m.filters }
func (m *mockPump) SetFilters(f *analytics.AnalyticFilters) { m.filters = f }
func (m *mockPump) GetTimeout() time.Duration               { return m.timeout }
func (m *mockPump) SetTimeout(d time.Duration)              { m.timeout = d }
func (m *mockPump) GetOmitDetailEnable() bool               { return m.omitDetail }
func (m *mockPump) SetOmitDetailEnable(b bool)              { m.omitDetail = b }

func TestFilterData_NoFilters(t *testing.T) {
	pump := &mockPump{
		filters: &analytics.AnalyticFilters{},
	}
	keys := []interface{}{
		analytics.AnalyticsRecode{Username: "alice", Reason: "test"},
		analytics.AnalyticsRecode{Username: "bob", Reason: "test2"},
	}

	result := filterData(pump, keys)
	assert.Len(t, result, 2)
	assert.Equal(t, keys, result)
}

func TestFilterData_OmitDetail(t *testing.T) {
	pump := &mockPump{
		omitDetail: true,
		filters:    &analytics.AnalyticFilters{},
	}
	keys := []interface{}{
		analytics.AnalyticsRecode{Username: "alice", Reason: "sensitive-info"},
	}

	result := filterData(pump, keys)
	assert.Len(t, result, 1)
	decoded := result[0].(analytics.AnalyticsRecode)
	assert.Empty(t, decoded.Reason, "Reason should be cleared when omitDetail is enabled")
}

func TestFilterData_SkipSpecificUsername(t *testing.T) {
	pump := &mockPump{
		filters: &analytics.AnalyticFilters{
			SkippedUsernames: []string{"bob"},
		},
	}
	keys := []interface{}{
		analytics.AnalyticsRecode{Username: "alice", Reason: "allow"},
		analytics.AnalyticsRecode{Username: "bob", Reason: "deny"},
		analytics.AnalyticsRecode{Username: "charlie", Reason: "allow"},
	}

	result := filterData(pump, keys)
	assert.Len(t, result, 2, "bob should be filtered out")
	assert.Equal(t, "alice", result[0].(analytics.AnalyticsRecode).Username)
	assert.Equal(t, "charlie", result[1].(analytics.AnalyticsRecode).Username)
}

func TestFilterData_OnlyAllowSpecificUsernames(t *testing.T) {
	pump := &mockPump{
		filters: &analytics.AnalyticFilters{
			Usernames: []string{"alice"},
		},
	}
	keys := []interface{}{
		analytics.AnalyticsRecode{Username: "alice", Reason: "allow"},
		analytics.AnalyticsRecode{Username: "bob", Reason: "deny"},
		analytics.AnalyticsRecode{Username: "charlie", Reason: "deny"},
	}

	result := filterData(pump, keys)
	assert.Len(t, result, 1, "only alice should remain")
	assert.Equal(t, "alice", result[0].(analytics.AnalyticsRecode).Username)
}

func TestFilterData_OmitDetailWithFilters(t *testing.T) {
	pump := &mockPump{
		omitDetail: true,
		filters: &analytics.AnalyticFilters{
			SkippedUsernames: []string{"bob"},
		},
	}
	keys := []interface{}{
		analytics.AnalyticsRecode{Username: "alice", Reason: "secret"},
		analytics.AnalyticsRecode{Username: "bob", Reason: "internal"},
	}

	result := filterData(pump, keys)
	assert.Len(t, result, 1)
	assert.Equal(t, "alice", result[0].(analytics.AnalyticsRecode).Username)
	assert.Empty(t, result[0].(analytics.AnalyticsRecode).Reason, "Reason should be cleared")
}

func TestFilterData_EmptyKeys(t *testing.T) {
	pump := &mockPump{
		filters: &analytics.AnalyticFilters{
			Usernames: []string{"alice"},
		},
		omitDetail: false,
	}
	keys := []interface{}{}

	result := filterData(pump, keys)
	assert.Empty(t, result)
}

func TestFilterData_AllFilteredOut(t *testing.T) {
	pump := &mockPump{
		filters: &analytics.AnalyticFilters{
			Usernames: []string{"admin"},
		},
	}
	keys := []interface{}{
		analytics.AnalyticsRecode{Username: "alice"},
		analytics.AnalyticsRecode{Username: "bob"},
	}

	result := filterData(pump, keys)
	assert.Empty(t, result)
}

func TestFilterData_AllSkipped(t *testing.T) {
	pump := &mockPump{
		filters: &analytics.AnalyticFilters{
			SkippedUsernames: []string{"alice", "bob"},
		},
	}
	keys := []interface{}{
		analytics.AnalyticsRecode{Username: "alice"},
		analytics.AnalyticsRecode{Username: "bob"},
	}

	result := filterData(pump, keys)
	assert.Empty(t, result)
}
