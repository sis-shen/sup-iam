package storage

import "time"

type AnalyticsStore interface {
	Connect() error
	AppendToSetPipelined(string, [][]byte) error
	SetExpire(string, time.Duration) error
	GetExpire(string) (time.Duration, error)
	SetKeyPrefix(string)
	WithStopChan(stopChan <-chan struct{}) AnalyticsStore
	WithExpireTime(time.Duration) AnalyticsStore
}

const (
	// AnalyticsKeyName defines the key name in redis which used to analytics.
	AnalyticsKeyName string = "iam-system-analytics"
)
