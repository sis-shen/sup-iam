package analytics

import (
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockStore implements storage.AnalyticsStore for testing
type mockStore struct {
	connectCalled      int32
	appendCalled       int32
	appendKey          string
	appendData         [][]byte
	appendReturnErr    bool
	connectReturnErr   bool
	mu                 sync.Mutex
	stopChan           <-chan struct{}
	expireTime         time.Duration
	keyPrefix          string
	setExpireCalled    bool
	getExpireCalled    bool
	getExpireReturn    time.Duration
	getExpireReturnErr error
	// flushNotify is signaled on each AppendToSetPipelined call for test synchronization
	flushNotify chan struct{}
}

func (m *mockStore) Connect() error {
	atomic.AddInt32(&m.connectCalled, 1)
	if m.connectReturnErr {
		return assert.AnError
	}
	return nil
}

func (m *mockStore) AppendToSetPipelined(key string, data [][]byte) error {
	atomic.AddInt32(&m.appendCalled, 1)
	m.mu.Lock()
	m.appendKey = key
	m.appendData = append(m.appendData, data...)
	m.mu.Unlock()
	// Notify test that a flush occurred (non-blocking)
	select {
	case m.flushNotify <- struct{}{}:
	default:
	}
	if m.appendReturnErr {
		return assert.AnError
	}
	return nil
}

func (m *mockStore) SetExpire(key string, d time.Duration) error {
	m.setExpireCalled = true
	return nil
}

func (m *mockStore) GetExpire(key string) (time.Duration, error) {
	m.getExpireCalled = true
	return m.getExpireReturn, m.getExpireReturnErr
}

func (m *mockStore) SetKeyPrefix(prefix string) {
	m.keyPrefix = prefix
}

func (m *mockStore) WithStopChan(stopChan <-chan struct{}) storage.AnalyticsStore {
	m.stopChan = stopChan
	return m
}

func (m *mockStore) WithExpireTime(d time.Duration) storage.AnalyticsStore {
	m.expireTime = d
	return m
}

// waitForFlush waits for a flush notification or times out.
func waitForFlush(t *testing.T, store *mockStore, timeout time.Duration) {
	t.Helper()
	select {
	case <-store.flushNotify:
	case <-time.After(timeout):
		t.Fatal("timeout waiting for flush")
	}
}

func TestNewAnalytics(t *testing.T) {
	store := &mockStore{}
	opts := &AnalyticsOptions{
		PoolSize:         4,
		RecordBufferSize: 100,
		FlushInterval:    100 * time.Millisecond,
	}

	a := NewAnalytics(opts, store)
	assert.NotNil(t, a)
	assert.Equal(t, 4, a.poolSize)
	assert.Equal(t, uint64(25), a.workerBuffSize) // 100/4 = 25
	assert.NotNil(t, a.recordChan)
	assert.Equal(t, 100, cap(a.recordChan))
	assert.Nil(t, a.stopChan, "stopChan should be nil before Start()")
}

func TestNewAnalytics_WorkerBufferSizeMinimum(t *testing.T) {
	store := &mockStore{}
	opts := &AnalyticsOptions{
		PoolSize:         10,
		RecordBufferSize: 5, // smaller than pool size
	}

	a := NewAnalytics(opts, store)
	assert.Equal(t, uint64(1), a.workerBuffSize, "workerBuffSize should be at least 1")
}

func TestAnalytics_Start_Stop(t *testing.T) {
	store := &mockStore{}
	store.flushNotify = make(chan struct{}, 10)
	opts := &AnalyticsOptions{
		PoolSize:         1,
		RecordBufferSize: 10,
		FlushInterval:    100 * time.Millisecond,
	}

	a := NewAnalytics(opts, store)
	err := a.Start()
	assert.NoError(t, err)
	assert.NotNil(t, a.stopChan)
	assert.Equal(t, int32(1), atomic.LoadInt32(&store.connectCalled))

	// Verify it's running
	err = a.RecordHit(&AnalyticsRecord{
		Username: "testuser",
		Resource: "doc:123",
		Action:   "read",
		Effect:   "allow",
	})
	assert.NoError(t, err)

	// Stop should not panic
	a.Stop()
}

func TestAnalytics_Start_Twice(t *testing.T) {
	store := &mockStore{}
	opts := NewAnalyticsOptions()
	opts.PoolSize = 1
	opts.RecordBufferSize = 10

	a := NewAnalytics(opts, store)
	err := a.Start()
	assert.NoError(t, err)

	err = a.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	a.Stop()
}

func TestAnalytics_Start_ConnectFailure(t *testing.T) {
	store := &mockStore{connectReturnErr: true}
	opts := NewAnalyticsOptions()

	a := NewAnalytics(opts, store)
	err := a.Start()
	assert.Error(t, err)
	assert.Nil(t, a.stopChan)
}

func TestAnalytics_RecordHit_AfterStop(t *testing.T) {
	store := &mockStore{}
	opts := &AnalyticsOptions{
		PoolSize:         1,
		RecordBufferSize: 10,
		FlushInterval:    100 * time.Millisecond,
	}

	a := NewAnalytics(opts, store)
	err := a.Start()
	require.NoError(t, err)
	a.Stop()

	err = a.RecordHit(&AnalyticsRecord{Username: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "analytics stopped")
}

func TestAnalytics_RecordHit_FullBuffer(t *testing.T) {
	store := &mockStore{}
	store.flushNotify = make(chan struct{}, 10)
	opts := &AnalyticsOptions{
		PoolSize:         1,
		RecordBufferSize: 2,
		FlushInterval:    1 * time.Hour, // long interval to not auto-flush
	}

	a := NewAnalytics(opts, store)
	err := a.Start()
	require.NoError(t, err)

	// Fill the buffer
	err = a.RecordHit(&AnalyticsRecord{Username: "user1", Resource: "res1"})
	assert.NoError(t, err)
	err = a.RecordHit(&AnalyticsRecord{Username: "user2", Resource: "res2"})
	assert.NoError(t, err)

	// Allow some time for worker to process
	waitForFlush(t, store, time.Second)

	a.Stop()
}

func TestAnalytics_WorkerFlushesOnStop(t *testing.T) {
	store := &mockStore{}
	store.flushNotify = make(chan struct{}, 10)
	opts := &AnalyticsOptions{
		PoolSize:         1,
		RecordBufferSize: 100,
		FlushInterval:    1 * time.Hour, // prevent auto-flush during test
	}

	a := NewAnalytics(opts, store)
	err := a.Start()
	require.NoError(t, err)

	// Record some hits
	for i := 0; i < 3; i++ {
		_ = a.RecordHit(&AnalyticsRecord{Username: "user", Resource: "res"})
	}

	// Stop triggers flush
	a.Stop()
	// 新版 analytics.go 在 stopChan 时 drain recordChan 中所有剩余数据后 flush
	// 3 条记录全部应被 flush
	store.mu.Lock()
	appendCount := len(store.appendData)
	store.mu.Unlock()
	assert.Equal(t, 3, appendCount, "All 3 records should have been flushed on stop")
}

func TestAnalytics_FlushOnBufferFull(t *testing.T) {
	store := &mockStore{}
	store.flushNotify = make(chan struct{}, 10)
	opts := &AnalyticsOptions{
		PoolSize:         1,
		RecordBufferSize: 100,
		FlushInterval:    1 * time.Hour, // long interval
	}

	a := NewAnalytics(opts, store)
	// Set workerBuffSize to 2 so it flushes every 2 records
	a.workerBuffSize = 2

	err := a.Start()
	require.NoError(t, err)

	// Send 2 records - should trigger flush
	// 第1条记录：buffer 只有1条，workerBuffSize=2，不会触发 flush
	_ = a.RecordHit(&AnalyticsRecord{Username: "user1", Resource: "res1"})
	// 第2条记录：buffer 达到2条 → 触发 flush
	_ = a.RecordHit(&AnalyticsRecord{Username: "user2", Resource: "res2"})
	waitForFlush(t, store, time.Second)

	a.Stop()

	store.mu.Lock()
	appendCount := store.appendCalled
	store.mu.Unlock()
	assert.Equal(t, int32(1), appendCount, "Should have flushed exactly once (buffer full with 2 records)")
}

func TestAnalytics_RecordHit_NilRecord(t *testing.T) {
	store := &mockStore{}
	store.flushNotify = make(chan struct{}, 10)
	opts := &AnalyticsOptions{
		PoolSize:         1,
		RecordBufferSize: 10,
		FlushInterval:    100 * time.Millisecond,
	}

	a := NewAnalytics(opts, store)
	err := a.Start()
	require.NoError(t, err)

	// Sending nil record should not panic
	err = a.RecordHit(nil)
	assert.NoError(t, err)

	a.Stop()
}

func TestAnalytics_MultipleWorkers(t *testing.T) {
	store := &mockStore{}
	store.flushNotify = make(chan struct{}, 10)
	opts := &AnalyticsOptions{
		PoolSize:         1,
		RecordBufferSize: 100,
		FlushInterval:    100 * time.Millisecond,
	}

	a := NewAnalytics(opts, store)
	err := a.Start()
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		_ = a.RecordHit(&AnalyticsRecord{
			Username: "user",
			Resource: "res",
			Action:   "read",
			Effect:   "allow",
		})
	}

	a.Stop()
	// Stop 后 worker 退出，数据应被 flush 到 store
	store.mu.Lock()
	appendCount := len(store.appendData)
	store.mu.Unlock()
	assert.GreaterOrEqual(t, appendCount, 10, "Should have flushed all 10 records on stop")
}
