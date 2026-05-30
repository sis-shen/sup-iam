package log

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type redisWriteSyncer struct {
	client *redis.Client
	key    string
	ctx    context.Context
}

var _ (zapcore.WriteSyncer) = (*redisWriteSyncer)(nil)

func (w *redisWriteSyncer) Write(p []byte) (n int, err error) {
	err = w.client.RPush(w.ctx, w.key, p).Err()
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *redisWriteSyncer) Sync() error { return nil }

// NewRedisCore 用已有的 Redis 客户端创建 zapcore.Core，可添加到现有 Logger
func NewRedisCore(client *redis.Client, key string, level Level) zapcore.Core {
	ws := &redisWriteSyncer{
		client: client,
		key:    key,
		ctx:    context.Background(),
	}
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	return zapcore.NewCore(encoder, ws, level)
}

// WrapWithRedis 包装现有 Logger，添加 Redis 输出（不改原 Logger 接口）
func WrapWithRedis(logger Logger, client *redis.Client, key string, level Level) Logger {
	redisCore := NewRedisCore(client, key, level)
	// 合并原有的 Core 和新的 Redis Core
	teeCore := zapcore.NewTee(logger.getZapLogger().Core(), redisCore)
	zapL := zap.New(teeCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return NewLogger(zapL)
}
