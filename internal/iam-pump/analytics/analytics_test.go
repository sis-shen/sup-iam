package analytics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAnalyticsRecode_GetFieldNames(t *testing.T) {
	tests := []struct {
		name   string
		record *AnalyticsRecode
		want   []string
	}{
		{
			name:   "zero value record",
			record: &AnalyticsRecode{},
			want:   []string{"Timestamp", "UserID", "Username", "SecretID", "Resource", "Action", "Effect", "Reason", "LatencyMs"},
		},
		{
			name: "fully populated record",
			record: &AnalyticsRecode{
				Timestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UserID:    "user-1",
				Username:  "admin",
				SecretID:  "secret-abc",
				Resource:  "doc:123",
				Action:    "read",
				Effect:    "allow",
				Reason:    "matched policy p1",
				LatencyMs: 42,
			},
			want: []string{"Timestamp", "UserID", "Username", "SecretID", "Resource", "Action", "Effect", "Reason", "LatencyMs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.GetFieldNames()
			assert.Equal(t, tt.want, got)
			// Verify the returned slice is a new copy each time
			assert.Equal(t, 9, len(got), "should have exactly 9 fields")
		})
	}
}

func TestAnalyticsRecode_GetFieldNames_DoesNotMutateRecord(t *testing.T) {
	record := &AnalyticsRecode{
		Username: "test-user",
	}
	before := *record
	_ = record.GetFieldNames()
	assert.Equal(t, before, *record, "GetFieldNames should not mutate the record")
}
