package analytics

import (
	"errors"
	"fmt"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/storage"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/vmihailenco/msgpack/v5"
	"sync"
	"time"
)

type AnalyticsRecord struct {
	Timestamp time.Time `json:"timestamp"`  // 请求时间
	UserID    string    `json:"user_id"`    // 用户标识
	Username  string    `json:"username"`   //用户名
	SecretID  string    `json:"secret_id"`  //用户提交的Secret标识
	Resource  string    `json:"resource"`   // 访问资源（如：doc:123）
	Action    string    `json:"action"`     // 操作（read/write/delete）
	Effect    string    `json:"effect"`     // allow/deny
	Reason    string    `json:"reason"`     // 决策原因（匹配到哪条策略/无权限）
	LatencyMs int64     `json:"latency_ms"` // 决策耗时（性能监控有价值）
}

type Analytics struct {
	store                 storage.AnalyticsStore
	poolSize              int
	recordChan            chan *AnalyticsRecord
	workerBuffSize        uint64
	bufferFlushInterval   time.Duration
	stopChan              chan struct{}
	poolWg                sync.WaitGroup
	enable                bool
	enableDetailRecording bool
	analyticsKeyName      string
}

const recordsBufferForcedFlushInterval time.Duration = time.Second

func NewAnalytics(options *AnalyticsOptions, store storage.AnalyticsStore) *Analytics {
	workerBufferSize := options.RecordBufferSize / options.PoolSize
	if workerBufferSize == 0 {
		workerBufferSize = 1
	}
	log.Debugf("worker buffer size: %d", workerBufferSize)
	return &Analytics{
		store:               store,
		poolSize:            options.PoolSize,
		recordChan:          make(chan *AnalyticsRecord, options.RecordBufferSize),
		workerBuffSize:      uint64(workerBufferSize),
		bufferFlushInterval: options.FlushInterval,
	}
}

func (r *Analytics) RecordHit(record *AnalyticsRecord) error {
	select {
	case <-r.stopChan:
		return errors.New("analytics stopped")
	default:
	}
	r.recordChan <- record
	return nil
}

func (r *Analytics) Start() error {
	if r.stopChan != nil {
		return fmt.Errorf("analytics has already started")
	}
	err := r.store.Connect()
	if err != nil {
		return err
	}
	r.stopChan = make(chan struct{})
	for i := 0; i < r.poolSize; i++ {
		r.poolWg.Add(1)
		go r.recordWorker()
	}
	return nil
}

func (r *Analytics) Stop() {
	if r.stopChan == nil {
		return
	}
	close(r.stopChan)

	r.poolWg.Wait()
}

func (r *Analytics) recordWorker() {
	defer r.poolWg.Done()
	// 预分配内存
	recordsBuffer := make([][]byte, 0, r.workerBuffSize)

	ticker := time.NewTicker(r.bufferFlushInterval)
	lastFlush := time.Now()
	for {
		var readyToSend bool
		select {
		case record, ok := <-r.recordChan:
			if !ok {
				//chan已经关闭了
				err := r.store.AppendToSetPipelined(r.analyticsKeyName, recordsBuffer)
				if err != nil {
					log.Errorf("Error AppendToSetPipelined", err.Error())
				}
				return
			}

			if encoded, err := msgpack.Marshal(record); err != nil {
				log.Errorf("Error msgpack.Marshal %s", err.Error())
			} else {
				recordsBuffer = append(recordsBuffer, encoded)
			}

			readyToSend = uint64(len(recordsBuffer)) == r.workerBuffSize
		case <-ticker.C:
			readyToSend = true
		case <-r.stopChan:
			//chan已经关闭了
			err := r.store.AppendToSetPipelined(r.analyticsKeyName, recordsBuffer)
			if err != nil {
				log.Errorf("Error AppendToSetPipelined", err.Error())
			}
			return
		}

		//+低流量保护
		if len(recordsBuffer) > 0 && (readyToSend || time.Since(lastFlush) > r.bufferFlushInterval) {
			err := r.store.AppendToSetPipelined(r.analyticsKeyName, recordsBuffer)
			lastFlush = time.Now()
			if err != nil {
				log.Errorf("Error AppendToSetPipelined", err.Error())
			}
			recordsBuffer = recordsBuffer[:0]
		}

	}
}
