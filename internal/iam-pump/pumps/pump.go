package pumps

import (
	"context"
)

type PumpInterface interface {
	GetName() string
	New() PumpInterface
	Init(interface{}) error
	WriteData(ctx context.Context, keyValues []interface{}) error
}
