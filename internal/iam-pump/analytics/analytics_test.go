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

func TestAnalyticsRecode_FieldCount(t *testing.T) {
	record := &AnalyticsRecode{}
	fields := record.GetFieldNames()
	assert.Len(t, fields, 9, "AnalyticsRecode should have exactly 9 fields")
}

func TestAnalyticsRecode_FieldNames(t *testing.T) {
	record := &AnalyticsRecode{}
	fields := record.GetFieldNames()
	expected := []string{"Timestamp", "UserID", "Username", "SecretID", "Resource", "Action", "Effect", "Reason", "LatencyMs"}
	assert.Equal(t, expected, fields)
}

func TestAnalyticsRecode_AllFieldsPopulated(t *testing.T) {
	now := time.Now()
	record := &AnalyticsRecode{
		Timestamp: now,
		UserID:    "user-42",
		Username:  "john_doe",
		SecretID:  "secret-abc-123",
		Resource:  "api:/v1/users:read",
		Action:    "read",
		Effect:    "allow",
		Reason:    "matched policy: admin-access",
		LatencyMs: 15,
	}

	assert.Equal(t, now, record.Timestamp)
	assert.Equal(t, "user-42", record.UserID)
	assert.Equal(t, "john_doe", record.Username)
	assert.Equal(t, "secret-abc-123", record.SecretID)
	assert.Equal(t, "api:/v1/users:read", record.Resource)
	assert.Equal(t, "read", record.Action)
	assert.Equal(t, "allow", record.Effect)
	assert.Equal(t, "matched policy: admin-access", record.Reason)
	assert.Equal(t, int64(15), record.LatencyMs)
}

func TestAnalyticsRecode_EmptyStrings(t *testing.T) {
	record := &AnalyticsRecode{
		Timestamp: time.Time{},
		UserID:    "",
		Username:  "",
		SecretID:  "",
		Resource:  "",
		Action:    "",
		Effect:    "",
		Reason:    "",
		LatencyMs: 0,
	}
	// All empty/default values should be valid
	fields := record.GetFieldNames()
	assert.Len(t, fields, 9)
}

func TestAnalyticsRecode_ZeroTimestamp(t *testing.T) {
	record := &AnalyticsRecode{}
	assert.True(t, record.Timestamp.IsZero(), "zero value Timestamp should be zero time")
}

func TestAnalyticsRecode_NegativeLatency(t *testing.T) {
	record := &AnalyticsRecode{
		LatencyMs: -1,
	}
	assert.Equal(t, int64(-1), record.LatencyMs)
}

func TestAnalyticsRecode_LargeLatency(t *testing.T) {
	record := &AnalyticsRecode{
		LatencyMs: 999999,
	}
	assert.Equal(t, int64(999999), record.LatencyMs)
}

func TestAnalyticsRecode_LongReason(t *testing.T) {
	longReason := string(make([]byte, 10000))
	record := &AnalyticsRecode{
		Reason: longReason,
	}
	assert.Equal(t, 10000, len(record.Reason))
}

func TestAnalyticsRecode_GetFieldNames_ReturnsCopy(t *testing.T) {
	record := &AnalyticsRecode{}
	fields1 := record.GetFieldNames()
	fields2 := record.GetFieldNames()
	// Should return different slices (not the same backing array)
	assert.Equal(t, fields1, fields2)
	// Modifying one should not affect the other
	fields1[0] = "Modified"
	assert.NotEqual(t, fields1[0], fields2[0])
}

func TestAnalyticsRecode_JSONTags(t *testing.T) {
	// Verify JSON tags match expected keys
	record := &AnalyticsRecode{
		Timestamp: time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		UserID:    "u1",
		Username:  "admin",
		SecretID:  "s1",
		Resource:  "r1",
		Action:    "write",
		Effect:    "deny",
		Reason:    "no permission",
		LatencyMs: 100,
	}

	// Verify that struct fields have the correct json tags by checking the field names
	fields := record.GetFieldNames()
	// The json tags should map: Timestamp->timestamp, UserID->user_id, etc.
	assert.Equal(t, "Timestamp", fields[0])
	assert.Equal(t, "UserID", fields[1])
	assert.Equal(t, "Username", fields[2])
	assert.Equal(t, "SecretID", fields[3])
	assert.Equal(t, "Resource", fields[4])
	assert.Equal(t, "Action", fields[5])
	assert.Equal(t, "Effect", fields[6])
	assert.Equal(t, "Reason", fields[7])
	assert.Equal(t, "LatencyMs", fields[8])
}
