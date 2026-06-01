// Package store defines storages which store the analytics data from iam-authz-server.
package storage

type AnalyticStoreInterface interface {
	Init(config interface{}) error
	GetName() string
	Connect() error
	GetAndDeleteSet(string) []interface{}
	SetKeyPrefix(string)
	WithShutDownChan(c chan bool) AnalyticStoreInterface
}

const (
	// AnalyticsKeyName defines the key name in redis which used to analytics.
	AnalyticsKeyName string = "iam-system-analytics"
)
