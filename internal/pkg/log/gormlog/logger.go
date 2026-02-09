package gormlog

import (
	"context"
	"fmt"
	zapLogger "github.com/sis-shen/sup-iam/internal/pkg/log"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"strings"
	"time"
)

type GormLog struct {
	zapLogger     zapLogger.Logger
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

func (g *GormLog) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *g
	newLogger.LogLevel = level
	return &newLogger
}

func (g *GormLog) Info(ctx context.Context, s string, i ...interface{}) {
	if g.LogLevel >= logger.Info {
		g.zapLogger.Infow(s, i)
	}
}

func (g *GormLog) Warn(ctx context.Context, s string, i ...interface{}) {
	if g.LogLevel >= logger.Warn {
		g.zapLogger.Warnw(s, i)
	}
}

func (g *GormLog) Error(ctx context.Context, s string, i ...interface{}) {
	if g.LogLevel >= logger.Error {
		g.zapLogger.Errorw(s, i)
	}
}

func (g *GormLog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Float64("elapsed_ms", float64(elapsed.Nanoseconds())/1e6),
		zap.Int64("rows", rows),
	}

	switch {
	case err != nil && g.LogLevel >= logger.Error:
		g.zapLogger.Error("SQL Error", append(fields, zap.Error(err))...)
	case elapsed > g.SlowThreshold && g.LogLevel >= logger.Warn:
		g.zapLogger.Warn("Slow Threshold", append(fields, zap.Duration("elapsed", elapsed), zap.Duration("threshold", g.SlowThreshold))...)
	case g.LogLevel >= logger.Info:
		g.zapLogger.Info("SQL Trace", fields...)
	}
}

func New(level logger.LogLevel, slowThreshold time.Duration) *GormLog {
	return &GormLog{
		zapLogger:     zapLogger.StdLogger(),
		LogLevel:      level,
		SlowThreshold: time.Duration(slowThreshold),
	}
}

func ParseLogLevel(s string) (logger.LogLevel, error) {
	switch strings.ToLower(s) {
	case "silent":
		return logger.Silent, nil
	case "error":
		return logger.Error, nil
	case "warn", "warning":
		return logger.Warn, nil
	case "info":
		return logger.Info, nil
	default:
		return logger.Info, fmt.Errorf("invalid gorm log level: %s", s)
	}
}

var _ logger.Interface = &GormLog{}
