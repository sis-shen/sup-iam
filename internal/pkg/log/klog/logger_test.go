package klog

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestZapLogger() *zap.Logger {
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&bytes.Buffer{}), zapcore.DebugLevel)
	return zap.New(core)
}

func TestInfoLogger_Write(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	logger := zap.New(core)

	w := &infoLogger{logger: logger}
	n, err := w.Write([]byte("test info message\n"))
	assert.NoError(t, err)
	assert.Equal(t, len("test info message\n"), n)
	assert.Contains(t, buf.String(), "test info message")
}

func TestWarnLogger_Write(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.WarnLevel)
	logger := zap.New(core)

	w := &warnLogger{logger: logger}
	n, err := w.Write([]byte("test warn message\n"))
	assert.NoError(t, err)
	assert.Equal(t, len("test warn message\n"), n)
	assert.Contains(t, buf.String(), "test warn message")
}

func TestErrorLogger_Write(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.ErrorLevel)
	logger := zap.New(core)

	w := &errorLogger{logger: logger}
	n, err := w.Write([]byte("test error message\n"))
	assert.NoError(t, err)
	assert.Equal(t, len("test error message\n"), n)
	assert.Contains(t, buf.String(), "test error message")
}

func TestInfoLogger_Write_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	logger := zap.New(core)

	w := &infoLogger{logger: logger}
	n, err := w.Write([]byte("\n"))
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestWarnLogger_Write_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.WarnLevel)
	logger := zap.New(core)

	w := &warnLogger{logger: logger}
	n, err := w.Write([]byte("\n"))
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestErrorLogger_Write_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.ErrorLevel)
	logger := zap.New(core)

	w := &errorLogger{logger: logger}
	n, err := w.Write([]byte("\n"))
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestInfoLogger_Write_MultiLine(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	logger := zap.New(core)

	w := &infoLogger{logger: logger}
	msg := "line1\nline2\n"
	n, err := w.Write([]byte(msg))
	assert.NoError(t, err)
	assert.Equal(t, len(msg), n)
	// Only the first line up to \n should be logged
	output := buf.String()
	assert.Contains(t, output, "line1")
}

func TestWriter_ImplementsWriteInterface(t *testing.T) {
	var _ interface{ Write([]byte) (int, error) } = &infoLogger{}
	var _ interface{ Write([]byte) (int, error) } = &warnLogger{}
	var _ interface{ Write([]byte) (int, error) } = &errorLogger{}
	var _ interface{ Write([]byte) (int, error) } = &fatalLogger{}
}
