package gormlog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

func TestNew(t *testing.T) {
	g := New(logger.Info, time.Second)
	assert.NotNil(t, g)
	assert.Equal(t, logger.Info, g.LogLevel)
	assert.Equal(t, time.Second, g.SlowThreshold)
	assert.NotNil(t, g.zapLogger)
}

func TestGormLog_ImplementsInterface(t *testing.T) {
	var _ logger.Interface = &GormLog{}
}

func TestGormLog_LogMode(t *testing.T) {
	g := &GormLog{
		zapLogger:     log.NewLogger(zap.NewNop()),
		LogLevel:      logger.Warn,
		SlowThreshold: time.Second,
	}

	updated := g.LogMode(logger.Info)
	assert.NotEqual(t, g, updated, "LogMode should return a different instance")
	assert.Equal(t, logger.Info, updated.(*GormLog).LogLevel)
	// Original should be unchanged
	assert.Equal(t, logger.Warn, g.LogLevel)
}

func TestGormLog_Info_AboveLevel(t *testing.T) {
	g := &GormLog{
		zapLogger: log.NewLogger(zap.NewNop()),
		LogLevel:  logger.Info,
	}
	// Info level >= Info, should not panic
	g.Info(context.Background(), "test message")
}

func TestGormLog_Info_BelowLevel(t *testing.T) {
	g := &GormLog{
		zapLogger: log.NewLogger(zap.NewNop()),
		LogLevel:  logger.Warn,
	}
	// Warn level < Info, Info should be suppressed
	g.Info(context.Background(), "test message") // should not panic
}

func TestGormLog_Warn_AboveLevel(t *testing.T) {
	g := &GormLog{
		zapLogger: log.NewLogger(zap.NewNop()),
		LogLevel:  logger.Warn,
	}
	g.Warn(context.Background(), "test warning")
}

func TestGormLog_Warn_BelowLevel(t *testing.T) {
	g := &GormLog{
		zapLogger: log.NewLogger(zap.NewNop()),
		LogLevel:  logger.Error,
	}
	g.Warn(context.Background(), "test warning") // should be suppressed
}

func TestGormLog_Error_AboveLevel(t *testing.T) {
	g := &GormLog{
		zapLogger: log.NewLogger(zap.NewNop()),
		LogLevel:  logger.Error,
	}
	g.Error(context.Background(), "test error")
}

func TestGormLog_Error_BelowLevel(t *testing.T) {
	g := &GormLog{
		zapLogger: log.NewLogger(zap.NewNop()),
		LogLevel:  logger.Silent,
	}
	g.Error(context.Background(), "test error") // should be suppressed
}

func TestGormLog_Trace_WithError(t *testing.T) {
	g := &GormLog{
		zapLogger:     log.NewLogger(zap.NewNop()),
		LogLevel:      logger.Error,
		SlowThreshold: time.Second,
	}
	begin := time.Now()
	fc := func() (string, int64) { return "SELECT 1", 1 }
	g.Trace(context.Background(), begin, fc, errors.New("db error"))
}

func TestGormLog_Trace_SlowQuery(t *testing.T) {
	g := &GormLog{
		zapLogger:     log.NewLogger(zap.NewNop()),
		LogLevel:      logger.Warn,
		SlowThreshold: time.Nanosecond,
	}
	begin := time.Now().Add(-time.Hour) // simulated slow query
	fc := func() (string, int64) { return "SELECT * FROM users", 100 }
	g.Trace(context.Background(), begin, fc, nil)
}

func TestGormLog_Trace_Normal(t *testing.T) {
	g := &GormLog{
		zapLogger:     log.NewLogger(zap.NewNop()),
		LogLevel:      logger.Info,
		SlowThreshold: time.Hour,
	}
	begin := time.Now()
	fc := func() (string, int64) { return "SELECT * FROM users", 5 }
	g.Trace(context.Background(), begin, fc, nil)
}

func TestGormLog_Trace_SilentLevel(t *testing.T) {
	g := &GormLog{
		zapLogger:     log.NewLogger(zap.NewNop()),
		LogLevel:      logger.Silent,
		SlowThreshold: time.Nanosecond,
	}
	begin := time.Now().Add(-time.Hour)
	fc := func() (string, int64) { return "SELECT * FROM users", 5 }
	// Silent level should suppress all logging including slow queries
	g.Trace(context.Background(), begin, fc, nil)
}

func TestGormLog_Trace_WithInfoLevelAndNoError(t *testing.T) {
	g := &GormLog{
		zapLogger:     log.NewLogger(zap.NewNop()),
		LogLevel:      logger.Info,
		SlowThreshold: time.Hour,
	}
	begin := time.Now()
	fc := func() (string, int64) { return "INSERT INTO logs (msg) VALUES ('test')", 1 }
	g.Trace(context.Background(), begin, fc, nil)
}

func TestGormLog_Info_WithArgs(t *testing.T) {
	g := &GormLog{
		zapLogger: log.NewLogger(zap.NewNop()),
		LogLevel:  logger.Info,
	}
	g.Info(context.Background(), "format %s %d", "arg1", 42)
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected logger.LogLevel
		wantErr  bool
	}{
		{"silent", logger.Silent, false},
		{"error", logger.Error, false},
		{"warn", logger.Warn, false},
		{"warning", logger.Warn, false},
		{"info", logger.Info, false},
		{"INFO", logger.Info, false},
		{"WARNING", logger.Warn, false},
		{"invalid", logger.Info, true},
		{"", logger.Info, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := ParseLogLevel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, logger.Info, level)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestGormLog_New_CustomLogger(t *testing.T) {
	zl := log.NewLogger(zap.NewNop())
	g := &GormLog{
		zapLogger:     zl,
		LogLevel:      logger.Info,
		SlowThreshold: 5 * time.Second,
	}
	assert.NotNil(t, g)
	g.Info(context.Background(), "custom logger test")
}
