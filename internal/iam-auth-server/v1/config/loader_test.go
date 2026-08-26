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

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, 8080, cfg.Grpc.Port)
	// 默认 TLS 关闭，向后兼容
	assert.False(t, cfg.Server.TLS.Enabled)
	assert.Empty(t, cfg.Server.TLS.CertFile)
	assert.Empty(t, cfg.Server.TLS.KeyFile)
}

func TestLoadConfigFile_Success(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 9999
grpc:
  port: 8080
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, 8080, cfg.Grpc.Port)
}

func TestLoadConfigFile_EmptyPath(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(origDir) })

	// Load 吞掉配置文件读取错误，返回默认配置
	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestLoadConfigFile_FileNotFound(t *testing.T) {
	// Load 吞掉配置文件读取错误，返回默认配置
	cfg, err := Load("/nonexistent/path/config.yaml")
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestLoadEnvVars_Bind(t *testing.T) {
	v := viper.New()
	// LoadEnvVars 不应 panic
	LoadEnvVars(v)
	// 设置环境变量后 viper 能正确读取
	t.Setenv("IAM_SERVER_MODE", "release")
	assert.Equal(t, "release", v.GetString("server.mode"))
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

func TestLoad_ServerTLSEnabled(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  tls:
    enabled: true
    cert_file: "/etc/iam/tls/server.crt"
    key_file: "/etc/iam/tls/server.key"
grpc:
  port: 9090
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.True(t, cfg.Server.TLS.Enabled)
	assert.Equal(t, "/etc/iam/tls/server.crt", cfg.Server.TLS.CertFile)
	assert.Equal(t, "/etc/iam/tls/server.key", cfg.Server.TLS.KeyFile)
}

func TestLoad_ServerTLSDisabled(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  tls:
    enabled: false
grpc:
  port: 9090
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.False(t, cfg.Server.TLS.Enabled)
	assert.Empty(t, cfg.Server.TLS.CertFile)
	assert.Empty(t, cfg.Server.TLS.KeyFile)
}

func TestLoad_ServerTLSDefaultDisabled(t *testing.T) {
	// 未配置 server.tls 段时，默认 enabled=false（向后兼容）
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
grpc:
  port: 9090
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.False(t, cfg.Server.TLS.Enabled)
	assert.Empty(t, cfg.Server.TLS.CertFile)
	assert.Empty(t, cfg.Server.TLS.KeyFile)
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
	cfg := NewConfig()
	cfg.Server.Port = 8080
	cfg.Grpc.Port = 99999

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gRPC端口无效")
}
