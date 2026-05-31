package log

import (
	"context"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew_CustomOptions(t *testing.T) {
	opts := NewOptions()
	opts.Level = "debug"
	opts.Format = "json"
	opts.OutputPaths = []string{"stdout"}
	opts.ErrorOutputPaths = []string{"stderr"}

	l := New(opts)
	assert.NotNil(t, l)
	assert.Equal(t, opts.Name, l.zapLogger.Name())
}

func TestNew_NilOptions(t *testing.T) {
	l := New(nil)
	assert.NotNil(t, l)
}

func TestNew_InvalidLevel(t *testing.T) {
	opts := NewOptions()
	opts.Level = "not-a-level"
	l := New(opts)
	assert.NotNil(t, l)
	// 无效级别回退到 InfoLevel
	assert.Equal(t, zap.InfoLevel, l.level)
}

func TestNew_ConsoleWithColor(t *testing.T) {
	opts := NewOptions()
	opts.Format = "console"
	opts.EnableColor = true
	l := New(opts)
	assert.NotNil(t, l)
}

func TestNewLogger_FromZapLogger(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	l := NewLogger(zapLogger)
	assert.NotNil(t, l)
}

func TestLogger_Info(t *testing.T) {
	l := New(NewOptions())
	l.Info("test message")
}

func TestLogger_Infof(t *testing.T) {
	l := New(NewOptions())
	l.Infof("test %s %d", "format", 42)
}

func TestLogger_Infow(t *testing.T) {
	l := New(NewOptions())
	l.Infow("test with keys", "key1", "value1", "key2", 2)
}

func TestLogger_Debug(t *testing.T) {
	opts := NewOptions()
	opts.Level = "debug"
	l := New(opts)
	l.Debug("debug message")
}

func TestLogger_Debugf(t *testing.T) {
	opts := NewOptions()
	opts.Level = "debug"
	l := New(opts)
	l.Debugf("debug %s", "format")
}

func TestLogger_Debugw(t *testing.T) {
	opts := NewOptions()
	opts.Level = "debug"
	l := New(opts)
	l.Debugw("debug with keys", "k", "v")
}

func TestLogger_Warn(t *testing.T) {
	l := New(NewOptions())
	l.Warn("warn message")
}

func TestLogger_Warnf(t *testing.T) {
	l := New(NewOptions())
	l.Warnf("warn %s", "format")
}

func TestLogger_Warnw(t *testing.T) {
	l := New(NewOptions())
	l.Warnw("warn with keys", "k", "v")
}

func TestLogger_Error(t *testing.T) {
	l := New(NewOptions())
	l.Error("error message")
}

func TestLogger_Errorf(t *testing.T) {
	l := New(NewOptions())
	l.Errorf("error %s", "format")
}

func TestLogger_Errorw(t *testing.T) {
	l := New(NewOptions())
	l.Errorw("error with keys", "k", "v")
}

func TestLogger_V_Enabled(t *testing.T) {
	opts := NewOptions()
	opts.V = 3
	l := New(opts)

	vLogger := l.V(2)
	assert.True(t, vLogger.Enabled())
}

func TestLogger_V_Disabled(t *testing.T) {
	opts := NewOptions()
	opts.V = 1
	l := New(opts)

	vLogger := l.V(5)
	assert.False(t, vLogger.Enabled())
}

func TestLogger_V_Info(t *testing.T) {
	opts := NewOptions()
	opts.V = 3
	l := New(opts)
	l.V(2).Info("v level info")
}

func TestLogger_Write(t *testing.T) {
	l := New(NewOptions())
	n, err := l.Write([]byte("write test"))
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
}

func TestLogger_WithValues(t *testing.T) {
	l := New(NewOptions())
	child := l.WithValues("addedKey", "addedValue")
	assert.NotNil(t, child)
}

func TestLogger_WithName(t *testing.T) {
	l := New(NewOptions())
	named := l.WithName("test-component")
	assert.NotNil(t, named)
}

func TestLogger_getZapLogger(t *testing.T) {
	l := New(NewOptions())
	zl := l.getZapLogger()
	assert.NotNil(t, zl)
}

func TestNoopInfoLogger(t *testing.T) {
	n := &noopInfoLogger{}
	assert.False(t, n.Enabled())
	n.Info("noop")
	n.Infof("noop %d", 1)
	n.Infow("noop", "k", "v")
}

// ---------------------------------------------------------------------------
// Global function tests (internal package, can access std)
// ---------------------------------------------------------------------------

func TestGlobal_V(t *testing.T) {
	opts := NewOptions()
	opts.Level = "debug"
	opts.V = 5
	Init(opts)

	vLogger := V(3)
	assert.NotNil(t, vLogger)
}

func TestGlobal_WithValues(t *testing.T) {
	opts := NewOptions()
	Init(opts)

	l := WithValues("key", "val")
	assert.NotNil(t, l)
}

func TestGlobal_WithName(t *testing.T) {
	opts := NewOptions()
	Init(opts)

	l := WithName("global-component")
	assert.NotNil(t, l)
}

func TestGlobal_Flush(t *testing.T) {
	opts := NewOptions()
	Init(opts)
	// 不应 panic
	Flush()
}

func TestGlobal_StdLogger(t *testing.T) {
	opts := NewOptions()
	Init(opts)

	l := StdLogger()
	assert.NotNil(t, l)

	sl := StdErrLogger()
	assert.NotNil(t, sl)

	sl2 := StdInfoLogger()
	assert.NotNil(t, sl2)
}

func TestGlobal_ZapLogger(t *testing.T) {
	opts := NewOptions()
	Init(opts)

	zl := ZapLogger()
	assert.NotNil(t, zl)
}

func TestGlobal_CheckIntLevel(t *testing.T) {
	opts := NewOptions()
	Init(opts)

	// Info level (-1*int32(zap.InfoLevel) in CheckIntLevel)
	enabled := CheckIntLevel(0)
	assert.True(t, enabled)
}

// ---------------------------------------------------------------------------
// Context tests
// ---------------------------------------------------------------------------

func TestWithContext_And_FromContext(t *testing.T) {
	opts := NewOptions()
	Init(opts)

	ctx := WithContext(context.Background())
	assert.NotNil(t, ctx)

	logger := FromContext(ctx)
	assert.NotNil(t, logger)
}

func TestFromContext_NilContext(t *testing.T) {
	logger := FromContext(nil)
	assert.NotNil(t, logger)
}

func TestFromContext_NoLoggerInContext(t *testing.T) {
	opts := NewOptions()
	Init(opts)

	logger := FromContext(context.Background())
	assert.NotNil(t, logger)
}

func TestLogger_L_WithContext(t *testing.T) {
	opts := NewOptions()
	Init(opts)

	ctx := context.WithValue(context.Background(), KeyRequestID, "req-123")
	ctx = context.WithValue(ctx, KeyUsername, "alice")

	l := std.L(ctx)
	assert.NotNil(t, l)
}

func TestLogger_clone(t *testing.T) {
	l := New(NewOptions())
	c := l.clone()
	assert.NotNil(t, c)
	// clone 返回新指针
	assert.True(t, l != c)
}

// ---------------------------------------------------------------------------
// HandleFields tests
// ---------------------------------------------------------------------------

func TestHandleFields_Empty(t *testing.T) {
	l, _ := zap.NewDevelopment()
	fields := handleFields(l, []interface{}{})
	assert.Empty(t, fields)
}

func TestHandleFields_OddNumberOfArgs(t *testing.T) {
	// 使用生产级别的 logger，避免 DPanic 触发 panic
	l, _ := zap.NewProduction()
	// 奇数个参数不应 panic
	_ = handleFields(l, []interface{}{"key1"})
}

// ---------------------------------------------------------------------------
// InfoLogger methods
// ---------------------------------------------------------------------------

func TestInfoLogger_Info(t *testing.T) {
	l := New(NewOptions())
	// 通过 V(0) 获取 infoLogger
	iLogger := l.V(0)
	if iLogger.Enabled() {
		iLogger.Infof("test %s", "format")
		iLogger.Infow("test", "k1", "v1")
	}
}

// ---------------------------------------------------------------------------
// Panic / Fatal coverage (only test that they can be called without crash)
// ---------------------------------------------------------------------------

func TestLogger_PanicAndFatal(t *testing.T) {
	// zap.Panic 始终会执行 panic，验证方法可调用
	l := New(NewOptions())
	assert.Panics(t, func() {
		l.Panic("panic message")
	})
	assert.Panics(t, func() {
		l.Panicf("panic %s", "format")
	})
	assert.Panics(t, func() {
		l.Panicw("panic with keys", "k", "v")
	})
}

// ---------------------------------------------------------------------------
// Options
// ---------------------------------------------------------------------------

func TestOptions_AddFlags(t *testing.T) {
	opts := NewOptions()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs)
	assert.NotNil(t, fs)

	err := fs.Parse([]string{
		"--log.level=warn",
		"--log.format=json",
		"--log.disable-caller=true",
		"--log.disable-stacktrace=true",
		"--log.enable-color=true",
		"--log.development=true",
		"--log.name=test-logger",
		"--log.v=2",
	})
	assert.NoError(t, err)
	assert.Equal(t, "warn", opts.Level)
	assert.Equal(t, "json", opts.Format)
	assert.True(t, opts.DisableCaller)
	assert.True(t, opts.DisableStacktrace)
	assert.True(t, opts.EnableColor)
	assert.True(t, opts.Development)
	assert.Equal(t, "test-logger", opts.Name)
	assert.Equal(t, int32(2), opts.V)
}

func TestOptions_String(t *testing.T) {
	opts := NewOptions()
	s := opts.String()
	assert.NotEmpty(t, s)
	assert.Contains(t, s, "level")
	assert.Contains(t, s, "format")
}

// ---------------------------------------------------------------------------
// Encoder functions
// ---------------------------------------------------------------------------

func TestTimeEncoder(t *testing.T) {
	// Verifies encoder function is callable (used internally by New())
	opts := NewOptions()
	opts.Format = "console"
	l := New(opts)
	assert.NotNil(t, l)
}
