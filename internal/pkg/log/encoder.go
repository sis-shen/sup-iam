package log

import (
	"time"

	"go.uber.org/zap/zapcore"
)

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339))
}

func milliSecondsDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(d.Milliseconds())
}
