package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestConfig(t *testing.T, path, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(path) })
}

func TestSetDefaults(t *testing.T) {
	v := viper.New()
	setDefaults(v)
	assert.Equal(t, "0.0.0.0", v.GetString("server.host"))
	assert.Equal(t, 8889, v.GetInt("server.port"))
	assert.Equal(t, "debug", v.GetString("server.mode"))
	assert.Equal(t, 9090, v.GetInt("grpc.port"))
	assert.Equal(t, "info", v.GetString("log.level"))
	assert.Equal(t, "console", v.GetString("log.format"))
}

func TestLoadConfigFile_Success(t *testing.T) {
	v := viper.New()
	setDefaults(v)

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 9999
grpc:
  port: 8080
`)

	err := LoadConfigFile(v, configPath)
	require.NoError(t, err)
	assert.Equal(t, 9999, v.GetInt("server.port"))
	assert.Equal(t, 8080, v.GetInt("grpc.port"))
}

func TestLoadConfigFile_EmptyPath(t *testing.T) {
	// 切换到临时目录以避免误用当前目录下的 config.yaml
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(origDir) })

	v := viper.New()
	err = LoadConfigFile(v, "")
	assert.Error(t, err)
}

func TestLoadConfigFile_FileNotFound(t *testing.T) {
	v := viper.New()
	err := LoadConfigFile(v, "/nonexistent/path/config.yaml")
	assert.Error(t, err)
}

func TestLoadEnvVars_Bind(t *testing.T) {
	v := viper.New()
	setDefaults(v)

	// LoadEnvVars 不应 panic
	LoadEnvVars(v)
	// 验证绑定成功
	assert.Equal(t, "debug", v.GetString("server.mode"))
}

func TestLoad_Success(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  host: "127.0.0.1"
  port: 7777
  mode: "test"
grpc:
  host: "127.0.0.1"
  port: 9999
log:
  level: "debug"
  format: "json"
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 7777, cfg.Server.Port)
	assert.Equal(t, "test", cfg.Server.Mode)
	assert.Equal(t, 9999, cfg.Grpc.Port)
}

func TestLoad_InvalidPort(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 99999
grpc:
  port: 9090
`)

	_, err := Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "端口无效")
}

func TestLoad_InvalidServerPort(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 0
grpc:
  port: 9090
`)

	_, err := Load(configPath)
	assert.Error(t, err)
}

func TestLoad_InvalidMode(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "invalid-mode"
grpc:
  port: 9090
`)

	_, err := Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的运行模式")
}

func TestAutoConfigFilePath_FromEnv(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "custom.yaml")
	writeTestConfig(t, configPath, `server:
  port: 5555
grpc:
  port: 9090
`)

	t.Setenv("IAM_CONFIG_FILE", configPath)
	path := autoConfigFilePath()
	assert.Equal(t, configPath, path)
}

func TestAutoConfigFilePath_ByEnv(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(origDir) })

	writeTestConfig(t, filepath.Join(tmpDir, "config.prod.yaml"), "server:\n  port: 8080\n")

	t.Setenv("IAM_ENV", "production")
	t.Setenv("GO_ENV", "")
	path := autoConfigFilePath()
	assert.Contains(t, path, "config.prod.yaml")
}

func TestAutoConfigFilePath_ByGoEnv(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(origDir) })

	writeTestConfig(t, filepath.Join(tmpDir, "config.dev.yaml"), "server:\n  port: 8080\n")

	t.Setenv("IAM_ENV", "")
	t.Setenv("GO_ENV", "development")
	path := autoConfigFilePath()
	assert.Contains(t, path, "config.dev.yaml")
}

func TestAutoConfigFilePath_Default(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(origDir) })

	writeTestConfig(t, filepath.Join(tmpDir, "config.yaml"), "server:\n  port: 8080\n")

	t.Setenv("IAM_ENV", "")
	t.Setenv("GO_ENV", "")
	path := autoConfigFilePath()
	assert.Contains(t, path, "config.yaml")
}

func TestValidateConfig_InvalidGRPCPort(t *testing.T) {
	v := viper.New()
	setDefaults(v)
	v.Set("server.port", 8080)
	v.Set("grpc.port", 99999)

	var cfg Config
	err := v.Unmarshal(&cfg)
	require.NoError(t, err)

	err = validateConfig(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gRPC端口无效")
}
