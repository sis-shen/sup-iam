package log_test

import (
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithName(t *testing.T) {
	defer log.Flush()
	logger := log.WithName("test")
	logger.Infow("hello world", "foo", "bar")
}

func TestWithValues(t *testing.T) {
	defer log.Flush()
	logger := log.WithValues("key", "value")

	logger.Info("hello world")
	logger.Info("hello world")
}

func TestV(t *testing.T) {
	defer log.Flush()

	opts := log.NewOptions()
	opts.V = 3
	logger := log.New(opts)
	logger.V(2).Info("hello world")
	logger.V(3).Info("hello world")
	logger.V(4).Info("you can't see me")
}

func Test_Option(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ExitOnError)
	opt := log.NewOptions()
	opt.AddFlags(fs)

	args := []string{"--log.level=debug"}
	err := fs.Parse(args)
	assert.Nil(t, err)

	assert.Equal(t, "debug", opt.Level)
}

func TestGlobalLogger(t *testing.T) {
	defer log.Flush()
	opts := log.NewOptions()
	opts.Level = "debug"
	log.Init(opts)

	log.Debug("hello world")
	log.Info("hello world")
	log.Warn("hello World")
	log.Error("hello world")
}
