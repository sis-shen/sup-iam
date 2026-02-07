package log_test

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/sis-shen/sup-iam/internal/pkg/log"
)

func TestOptions_Validate(t *testing.T) {
	opts := &log.Options{
		Level:            "test",
		Format:           "test",
		EnableColor:      true,
		DisableCaller:    false,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	errs := opts.Validate()
	expected := `[unrecognized level: "test" not a valid log format: "test"]`
	assert.Equal(t, expected, fmt.Sprintf("%s", errs))
}

func TestOptions_Defaults_Work(t *testing.T) {
	opts := log.NewOptions()

	if errs := opts.Validate(); len(errs) != 0 {
		t.Fatalf("default options should be valid, got errors: %v", errs)
	}

	if err := opts.Build(); err != nil {
		t.Fatalf("Build() failed with default options: %v", err)
	}
}

func TestOptions_Flags_Parse(t *testing.T) {
	opts := log.NewOptions()

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs)

	args := []string{
		"--log.level=debug",
		"--log.format=json",
		"--log.enable-color=true",
		"--log.output-paths=stdout",
		"--log.error-output-paths=stderr",
		"--log.name=blackbox-test",
	}

	if err := fs.Parse(args); err != nil {
		t.Fatalf("flag parse failed: %v", err)
	}

	if errs := opts.Validate(); len(errs) != 0 {
		t.Fatalf("options should be valid after flag parse, got errors: %v", errs)
	}

	if err := opts.Build(); err != nil {
		t.Fatalf("Build() failed after flag parse: %v", err)
	}
}

func TestOptions_Invalid_Config_Rejected(t *testing.T) {
	opts := log.NewOptions()

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs)

	args := []string{
		"--log.level=not-a-level",
		"--log.format=xml",
	}

	if err := fs.Parse(args); err != nil {
		t.Fatalf("flag parse failed: %v", err)
	}

	errs := opts.Validate()
	if len(errs) == 0 {
		t.Fatalf("expected validation errors, got none")
	}
}

func TestOptions_String_NotEmpty(t *testing.T) {
	opts := log.NewOptions()

	s := opts.String()
	if s == "" {
		t.Fatalf("String() should not return empty string")
	}
}

func TestOptions_Multiple_Formats(t *testing.T) {
	cases := []struct {
		name   string
		level  string
		format string
	}{
		{"console-info", "info", "console"},
		{"json-debug", "debug", "json"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts := log.NewOptions()

			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			opts.AddFlags(fs)

			args := []string{
				"--log.level=" + c.level,
				"--log.format=" + c.format,
			}

			if err := fs.Parse(args); err != nil {
				t.Fatalf("flag parse failed: %v", err)
			}

			if errs := opts.Validate(); len(errs) != 0 {
				t.Fatalf("unexpected validation errors: %v", errs)
			}

			if err := opts.Build(); err != nil {
				t.Fatalf("Build() failed: %v", err)
			}
		})
	}
}
