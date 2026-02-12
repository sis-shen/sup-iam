package dbmysql

import (
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"gorm.io/gorm"
	"time"
)

const (
	callBackBeforeName = "core:before"
	callBackAfterName  = "core:after"
	startTime          = "_start_time"
)

// TracePlugin defines gorm plugin used to trace sql.
type TracePlugin struct{}

// Name returns the name of trace plugin.
func (op *TracePlugin) Name() string {
	return "tracePlugin"
}

var _ gorm.Plugin = &TracePlugin{}

func (op *TracePlugin) Initialize(db *gorm.DB) error {
	//开始前
	_ = db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, before)
	_ = db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, before)
	_ = db.Callback().Delete().Before("gorm:delete").Register(callBackBeforeName, before)
	_ = db.Callback().Update().Before("gorm:update").Register(callBackBeforeName, before)
	_ = db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, before)
	_ = db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, before)

	//结束后
	_ = db.Callback().Create().After("gorm:after").Register(callBackAfterName, after)
	_ = db.Callback().Query().After("gorm:after").Register(callBackAfterName, after)
	_ = db.Callback().Delete().After("gorm:after").Register(callBackAfterName, after)
	_ = db.Callback().Update().After("gorm:after").Register(callBackAfterName, after)
	_ = db.Callback().Row().After("gorm:row").Register(callBackAfterName, after)
	_ = db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after)

	return nil
}

func before(db *gorm.DB) {
	db.InstanceSet(startTime, time.Now())
}

func after(db *gorm.DB) {
	_ts, isExist := db.InstanceGet(startTime)
	if !isExist {
		return
	}
	ts, ok := _ts.(time.Time)
	if !ok {
		return
	}
	log.Infof("sql cost time: %fs", time.Since(ts).Seconds())
}
