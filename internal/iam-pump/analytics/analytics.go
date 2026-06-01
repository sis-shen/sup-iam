package analytics

import (
	"reflect"
	"time"
)

type AnalyticsRecode struct {
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

func (a *AnalyticsRecode) GetFieldNames() []string {
	val := reflect.ValueOf(a).Elem()
	fields := make([]string, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		fields[i] = val.Type().Field(i).Name
	}
	return fields
}
