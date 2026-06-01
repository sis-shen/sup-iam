package pumps

import (
	"context"
	"errors"
	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"time"
)

type PumpInterface interface {
	GetName() string
	New() PumpInterface
	Init(interface{}) error
	WriteData(ctx context.Context, keyValues []interface{}) error
	GetFilters() *analytics.AnalyticFilters
	SetFilters(*analytics.AnalyticFilters)
	GetTimeout() time.Duration
	SetTimeout(time.Duration)
	GetOmitDetailEnable() bool
	SetOmitDetailEnable(bool)
}

var availablePumps = map[string]PumpInterface{}

func GetPumpByName(name string) (PumpInterface, error) {
	if len(availablePumps) == 0 {
		initProtoType()
	}
	if pump, ok := availablePumps[name]; ok && pump != nil {
		return pump.New(), nil
	}

	return nil, errors.New(name + " Not found")
}

// 原型模式
func initProtoType() {
	availablePumps = map[string]PumpInterface{
		"mongo": &MongoPump{},
	}
}
