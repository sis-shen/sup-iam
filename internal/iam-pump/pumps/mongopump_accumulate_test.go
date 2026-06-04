package pumps

import (
	"testing"
	"time"

	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"github.com/stretchr/testify/assert"
)

func TestAccumulateSet_Empty(t *testing.T) {
	m := &MongoPump{
		config: &MongoConf{
			MaxInsertBatchSizeBytes: 10 * MiB,
			MaxDocumentSizeBytes:    10 * MiB,
		},
	}

	result := m.accumulateSet([]interface{}{})
	assert.Empty(t, result)
}

func TestAccumulateSet_SingleItem(t *testing.T) {
	m := &MongoPump{
		config: &MongoConf{
			MaxInsertBatchSizeBytes: 10 * MiB,
			MaxDocumentSizeBytes:    10 * MiB,
		},
	}

	record := analytics.AnalyticsRecode{
		Timestamp: time.Now(),
		UserID:    "user-1",
		Username:  "testuser",
		SecretID:  "secret-1",
		Resource:  "doc:123",
		Action:    "read",
		Effect:    "allow",
		Reason:    "matched policy",
		LatencyMs: 42,
	}

	result := m.accumulateSet([]interface{}{record})
	assert.Len(t, result, 1)
	assert.Len(t, result[0], 1)
}

func TestAccumulateSet_MultipleItems(t *testing.T) {
	m := &MongoPump{
		config: &MongoConf{
			MaxInsertBatchSizeBytes: 10 * MiB,
			MaxDocumentSizeBytes:    10 * MiB,
		},
	}

	items := make([]interface{}, 5)
	for i := 0; i < 5; i++ {
		items[i] = analytics.AnalyticsRecode{
			SecretID:  "secret-1",
			Reason:    "policy matched",
			LatencyMs: int64(i),
		}
	}

	result := m.accumulateSet(items)
	assert.Len(t, result, 1)
	assert.Len(t, result[0], 5)
}

func TestAccumulateSet_DocumentTooLarge(t *testing.T) {
	m := &MongoPump{
		config: &MongoConf{
			MaxInsertBatchSizeBytes: 10 * MiB,
			MaxDocumentSizeBytes:    100,
		},
	}

	// Reason is long enough to exceed the small MaxDocumentSizeBytes
	record := analytics.AnalyticsRecode{
		SecretID: "secret-1",
		Reason:   string(make([]byte, 2000)),
	}

	result := m.accumulateSet([]interface{}{record})
	assert.Len(t, result, 1)
	assert.Len(t, result[0], 1)
}

func TestAccumulateSet_MultipleBatches(t *testing.T) {
	m := &MongoPump{
		config: &MongoConf{
			MaxInsertBatchSizeBytes: 3000,
			MaxDocumentSizeBytes:    10 * MiB,
		},
	}

	// Each item costs about len(reason) + 1024 bytes = 1124 bytes
	// With batch size 3000: item0+item1 = 2248 < 3000, item0+item1+item2 = 3372 > 3000
	// So items 0-1 go in first batch, item 2 goes in second batch
	items := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		items[i] = analytics.AnalyticsRecode{
			SecretID:  "secret-1",
			Reason:    string(make([]byte, 100)),
			LatencyMs: int64(i),
		}
	}

	result := m.accumulateSet(items)
	assert.Len(t, result, 2, "Should create 2 batches")
	assert.Len(t, result[0], 2, "First batch should have 2 items")
	assert.Len(t, result[1], 1, "Second batch should have 1 item")
}

func TestAccumulateSet_ZeroMaxDocumentSize(t *testing.T) {
	m := &MongoPump{
		config: &MongoConf{
			MaxInsertBatchSizeBytes: 10 * MiB,
			MaxDocumentSizeBytes:    0,
		},
	}

	record := analytics.AnalyticsRecode{
		Reason: "test",
	}

	result := m.accumulateSet([]interface{}{record})
	assert.Len(t, result, 1)
	assert.Len(t, result[0], 1)
}

func TestAccumulateSet_ItemsExceedBatchButNotSplit(t *testing.T) {
	m := &MongoPump{
		config: &MongoConf{
			MaxInsertBatchSizeBytes: 3000,
			MaxDocumentSizeBytes:    10 * MiB,
		},
	}

	// Each item has reason of 500 bytes + 1024 overhead = ~1524 bytes
	// 2 items would be ~3048 bytes, exceeding 3000
	items := make([]interface{}, 2)
	for i := 0; i < 2; i++ {
		items[i] = analytics.AnalyticsRecode{
			SecretID: "secret-1",
			Reason:   string(make([]byte, 500)),
		}
	}

	result := m.accumulateSet(items)
	// First batch gets 1 item (since second would exceed), second batch gets the other
	assert.Len(t, result, 2)
	assert.Len(t, result[0], 1)
	assert.Len(t, result[1], 1)
}

func TestAccumulateSet_SingleItemExceedsBatch(t *testing.T) {
	m := &MongoPump{
		config: &MongoConf{
			MaxInsertBatchSizeBytes: 100,
			MaxDocumentSizeBytes:    10 * MiB,
		},
	}

	record := analytics.AnalyticsRecode{
		Reason: "test",
	}

	result := m.accumulateSet([]interface{}{record})
	assert.Len(t, result, 1, "Single item should still be included even if it exceeds batch size")
	assert.Len(t, result[0], 1)
}
