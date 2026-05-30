package log

import (
	"context"
	"fmt"
	"github.com/sis-shen/sup-iam/internal/pkg/log/klog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"sync"
)

// InfoLogger represents the ability to log non-error messages, at a particular verbosity.
type InfoLogger interface {
	// Info logs a non-error message with the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Info(msg string, fields ...Field)
	Infof(format string, v ...interface{})
	Infow(msg string, keysAndValues ...interface{})

	// Enabled tests whether this InfoLogger is enabled.  For example,
	// commandline flags might be used to set the logging verbosity and disable
	// some info logs.
	Enabled() bool
}

// Logger represents the ability to log messages, both errors and not.
type Logger interface {
	// All Loggers implement InfoLogger.  Calling InfoLogger methods directly on
	// a Logger value is equivalent to calling them on a V(0) InfoLogger.  For
	// example, logger.Info() produces the same result as logger.V(0).Info.
	InfoLogger

	Debug(msg string, fields ...Field)
	Debugf(format string, v ...interface{})
	Debugw(msg string, keysAndValues ...interface{})

	Warn(msg string, fields ...Field)
	Warnf(format string, v ...interface{})
	Warnw(msg string, keysAndValues ...interface{})

	Error(msg string, fields ...Field)
	Errorf(format string, v ...interface{})
	Errorw(msg string, keysAndValues ...interface{})

	Panic(msg string, fields ...Field)
	Panicf(format string, v ...interface{})
	Panicw(msg string, keysAndValues ...interface{})

	Fatal(msg string, fields ...Field)
	Fatalf(format string, v ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	// V returns an InfoLogger value for a specific verbosity level.  A higher
	// verbosity level means a log message is less important.  It's illegal to
	// pass a log level less than zero.
	V(level int32) InfoLogger

	Write(p []byte) (n int, err error)

	// WithValues adds some key-value pairs of context to a logger.
	// See Info for documentation on how key/value pairs work.
	WithValues(keysAndValues ...interface{}) Logger

	// WithName adds a new element to the logger's name.
	// Successive calls with WithName continue to append
	// suffixes to the logger's name.  It's strongly recommended
	// that name segments contain only letters, digits, and hyphens
	// (see the package documentation for more information).
	WithName(name string) Logger

	// WithContext returns a copy of context in which the log value is set.
	WithContext(ctx context.Context) context.Context

	// Flush calls the underlying Core's Sync method, flushing any buffered
	// log entries. Applications should take care to call Sync before exiting.
	Flush()

	getZapLogger() *zap.Logger
}

var _ Logger = &zapLogger{}

// noopInfoLogger is a logr.InfoLogger that's always disabled, and does nothing.
type noopInfoLogger struct{}

func (l *noopInfoLogger) Enabled() bool                    { return false }
func (l *noopInfoLogger) Info(_ string, _ ...Field)        {}
func (l *noopInfoLogger) Infof(_ string, _ ...interface{}) {}
func (l *noopInfoLogger) Infow(_ string, _ ...interface{}) {}

var disabledInfoLogger = &noopInfoLogger{}

// infoLogger is a logger.InfoLogger that uses Zap to log at a particular
// level.  The level has already been converted to a Zap level. But we
// use klog typed logLevel ,which is to say that `logLevel = -1*zapLevel`.

type infoLogger struct {
	vlevel int32
	level  zapcore.Level
	log    *zap.Logger
}

func (l *infoLogger) Enabled() bool { return true }

func (l *infoLogger) Info(msg string, fields ...Field) {
	if checkEntry := l.log.Check(l.level, msg); checkEntry != nil {
		checkEntry.Write(fields...)
	}
}

func (l *infoLogger) Infof(format string, v ...interface{}) {
	if checkEntry := l.log.Check(l.level, fmt.Sprintf(format, v...)); checkEntry != nil {
		checkEntry.Write()
	}
}

func (l *infoLogger) Infow(msg string, keysAndValues ...interface{}) {
	if checkEntry := l.log.Check(l.level, msg); checkEntry != nil {
		checkEntry.Write(handleFields(l.log, keysAndValues)...)
	}
}

// handleFields converts a bunch of arbitrary key-value pairs into Zap fields.  It takes
// additional pre-converted Zap fields, for use with automatically attached fields, like
// `error`.
func handleFields(l *zap.Logger, args []interface{}, additional ...zap.Field) []zap.Field {
	if len(args) == 0 {
		return additional
	}

	// unlike Zap, we can be pretty sure users aren't passing structured
	// fields (since logr has no concept of that), so guess that we need a
	// little less space.
	fields := make([]zap.Field, 0, len(args)/2+len(additional))
	for i := 0; i < len(args); {
		// check just in case for strongly-typed Zap fields, which is illegal (since
		// it breaks implementation agnosticism), so we can give a better error message.
		if _, ok := args[i].(zap.Field); ok {
			l.DPanic("strongly-typed Zap Field pass to logr", zap.Any("field", args[i]))
			break
		}

		// make sure this isn't a mismatched key
		if i == len(args)-1 {
			l.DPanic("odd number of arguments passed as key-value pairs for logging", zap.Any("ignored key", args[i]))

			break
		}

		// process a key-value pair,
		// ensuring that the key is a string
		key, val := args[i], args[i+1]
		keyStr, isString := key.(string)
		if !isString {
			l.DPanic(
				"non-string key argument passed to logging, ignoring all later arguments",
				zap.Any("invalid key", key),
			)
			break
		}
		fields = append(fields, zap.Any(keyStr, val))
		i += 2
	}
	return append(fields, additional...)
}

var (
	std = New(NewOptions())
	mu  sync.Mutex
)

func Init(opts *Options) {
	mu.Lock()
	defer mu.Unlock()
	std = New(opts)
}

type zapLogger struct {
	zapLogger *zap.Logger
	infoLogger
}

// NewLogger create Logger from zap.Logger
func NewLogger(l *zap.Logger) *zapLogger {
	return &zapLogger{
		zapLogger: l,
		infoLogger: infoLogger{
			log:    l,
			level:  zap.InfoLevel,
			vlevel: 0,
		},
	}
}

func (l *zapLogger) getZapLogger() *zap.Logger {
	return l.zapLogger
}

// New create logger by opts which can custmoized by command arguments.
func New(opts *Options) *zapLogger {
	if opts == nil {
		opts = NewOptions()
	}

	// 如果出错，则默认Info级别
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	encodeLevel := zapcore.CapitalLevelEncoder
	// 只有使用控制台输出时才能开启颜色输出
	if opts.Format == consoleFormat && opts.EnableColor {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encodeLevel,
		EncodeTime:     timeEncoder,
		EncodeDuration: milliSecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	loggerConfig := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       opts.Development,
		DisableCaller:     opts.DisableCaller,
		DisableStacktrace: opts.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         opts.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      opts.OutputPaths,
		ErrorOutputPaths: opts.ErrorOutputPaths,
	}

	var err error
	l, err := loggerConfig.Build(zap.AddStacktrace(zapcore.PanicLevel), zap.AddCallerSkip(1))

	if err != nil {
		panic(err)
	}
	logger := &zapLogger{
		zapLogger: l.Named(opts.Name),
		infoLogger: infoLogger{
			level:  zapLevel,
			log:    l,
			vlevel: opts.V,
		},
	}

	klog.InitLogger(l)
	zap.RedirectStdLog(l)
	return logger
}

// StdLogger returns global std logger. 保留一个获取zaplogger的接口
func StdLogger() *zapLogger {
	return std
}

// StdErrLogger returns logger of standard library which writes to supplied zap
// logger at error level.
func StdErrLogger() *log.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.zapLogger, zapcore.ErrorLevel); err == nil {
		return l
	}

	return nil
}

// StdInfoLogger returns logger of standard library which writes to supplied zap
// logger at info level.
func StdInfoLogger() *log.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.zapLogger, zapcore.InfoLevel); err == nil {
		return l
	}

	return nil
}

func ZapLogger() *zap.Logger {
	return std.zapLogger
}

func V(level int32) InfoLogger { return std.V(level) }

// V 返回有V Level决定是否生效的日志器,传入的是Klog格式的正数level，因此zapLevel = kLevel * -1
func (l *zapLogger) V(level int32) InfoLogger {
	if l.zapLogger.Core().Enabled(zapcore.InfoLevel) && level <= l.vlevel {
		return &infoLogger{
			level:  zapcore.InfoLevel,
			vlevel: level,
			log:    l.zapLogger,
		}
	}

	return disabledInfoLogger
}
func (l *zapLogger) Write(p []byte) (n int, err error) {
	l.zapLogger.Info(string(p))
	return len(p), nil
}

// WithValues creates a child logger and adds Zap fields to it.
func WithValues(keysAndValues ...interface{}) Logger { return std.WithValues(keysAndValues...) }

func (l *zapLogger) WithValues(keysAndValues ...interface{}) Logger {
	newLogger := l.zapLogger.With(handleFields(l.zapLogger, keysAndValues)...)

	return NewLogger(newLogger)
}

// WithName adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func WithName(s string) Logger { return std.WithName(s) }

func (l *zapLogger) WithName(s string) Logger {
	newLogger := l.zapLogger.Named(s)
	return NewLogger(newLogger)
}

// Flush calls the underlying Core's Sync method, flushing any buffered
// log entries. Applications should take care to call Sync before exiting.
func Flush() { std.Flush() }

func (l *zapLogger) Flush() {
	_ = l.zapLogger.Sync()
}

// CheckIntLevel used for other log wrapper such as klog which return if logging a
// message at the specified level is enabled.
func CheckIntLevel(level int32) bool {
	zapLevel := zapcore.Level(-1 * level)
	return std.zapLogger.Core().Enabled(zapLevel)
}

func Debug(msg string, fields ...Field) {
	std.zapLogger.Debug(msg, fields...)
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.zapLogger.Debug(msg, fields...)
}

func Debugf(format string, args ...interface{}) {
	std.zapLogger.Sugar().Debugf(format, args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Debugf(format, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Debugw(msg, keysAndValues...)
}

func Info(msg string, fields ...Field) {
	std.zapLogger.Info(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.zapLogger.Info(msg, fields...)
}

func Infof(format string, args ...interface{}) {
	std.zapLogger.Sugar().Infof(format, args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	std.zapLogger.Sugar().Infof(format, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Infow(msg, keysAndValues...)
}

func (l *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Infow(msg, keysAndValues...)
}

func Warn(msg string, fields ...Field) {
	std.zapLogger.Warn(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.zapLogger.Warn(msg, fields...)
}

func Warnf(format string, args ...interface{}) {
	std.zapLogger.Sugar().Warnf(format, args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Warnf(format, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Warnw(msg, keysAndValues...)
}

func (l *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Warnw(msg, keysAndValues...)
}

func Error(msg string, fields ...Field) {
	std.zapLogger.Error(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.zapLogger.Error(msg, fields...)
}

func Errorf(format string, args ...interface{}) {
	std.zapLogger.Sugar().Errorf(format, args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Errorf(format, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Errorw(msg, keysAndValues...)
}

func Panic(msg string, fields ...Field) {
	std.zapLogger.Panic(msg, fields...)
}

func (l *zapLogger) Panic(msg string, fields ...Field) {
	l.zapLogger.Panic(msg, fields...)
}

func Panicf(format string, args ...interface{}) {
	std.zapLogger.Sugar().Panicf(format, args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Panicf(format, args...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Panicw(msg, keysAndValues...)
}

func (l *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Panicw(msg, keysAndValues...)
}

func Fatal(msg string, fields ...Field) {
	std.zapLogger.Fatal(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.zapLogger.Fatal(msg, fields...)
}

func Fatalf(format string, args ...interface{}) {
	std.zapLogger.Sugar().Fatalf(format, args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Fatalf(format, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Fatalw(msg, keysAndValues...)
}

// L method output with specified context value.
func L(ctx context.Context) *zapLogger {
	return std.L(ctx)
}

func (l *zapLogger) L(ctx context.Context) *zapLogger {
	lg := l.clone()

	requestID, _ := ctx.Value(KeyRequestID).(string)
	username, _ := ctx.Value(KeyUsername).(string)
	lg.zapLogger = lg.zapLogger.With(zap.String(KeyRequestID, requestID), zap.String(KeyUsername, username))

	return lg
}

func (l *zapLogger) clone() *zapLogger {
	cpy := *l

	return &cpy
}
