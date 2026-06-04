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
	require.NotNil(t, cfg.Server)
	require.NotNil(t, cfg.JWT)
	require.NotNil(t, cfg.MySQL)
	require.NotNil(t, cfg.Redis)
	require.NotNil(t, cfg.Log)

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, 3306, cfg.MySQL.Port)
	assert.Equal(t, 6379, cfg.Redis.Port)
}

func TestLoadConfigFile_Success(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 5555
jwt:
  secret_key: "test-jwt-secret"
mysql:
  password: "test-pass"
  port: 3307
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, 5555, cfg.Server.Port)
	assert.Equal(t, "test-jwt-secret", cfg.JWT.SecretKey)
	assert.Equal(t, 3307, cfg.MySQL.Port)
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
	assert.Equal(t, 9000, cfg.Server.Port)
}

func TestLoadConfigFile_FileNotFound(t *testing.T) {
	// Load 吞掉配置文件读取错误，返回默认配置
	cfg, err := Load("/nonexistent/config.yaml")
	require.NoError(t, err)
	assert.Equal(t, 9000, cfg.Server.Port)
}

func TestLoadEnvVars_Bind(t *testing.T) {
	v := viper.New()
	LoadEnvVars(v)
	// LoadEnvVars 不应 panic
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
jwt:
  secret_key: "my-jwt-secret"
mysql:
  password: "mysql-pass"
  port: 3306
redis:
  port: 6379
grpc:
  port: 9090
log:
  level: "debug"
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 7777, cfg.Server.Port)
	assert.Equal(t, "test", cfg.Server.Mode)
	assert.Equal(t, "my-jwt-secret", cfg.JWT.SecretKey)
	assert.Equal(t, "mysql-pass", cfg.MySQL.Password)
}

func TestLoad_NoJWTSecret(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "test"
jwt:
  secret_key: ""
mysql:
  password: "pass"
`)

	_, err := Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT密钥不能为空")
}

func TestLoad_NoMySQLPassword(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "test"
jwt:
  secret_key: "secret"
mysql:
  password: ""
`)

	_, err := Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MySQL密码不能为空")
}

func TestLoad_InvalidServerPort(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 99999
  mode: "test"
jwt:
  secret_key: "secret"
mysql:
  password: "pass"
`)

	_, err := Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "服务器端口无效")
}

func TestLoad_InvalidMySQLPort(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "test"
jwt:
  secret_key: "secret"
mysql:
  password: "pass"
  port: -1
`)

	_, err := Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MySQL端口无效")
}

func TestLoad_InvalidMode(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "bad-mode"
jwt:
  secret_key: "secret"
mysql:
  password: "pass"
`)

	_, err := Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的运行模式")
}

func TestAutoConfigFilePath_FromEnv(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "custom.yaml")
	writeTestConfig(t, configPath, `server:
  port: 5000
jwt:
  secret_key: "s"
mysql:
  password: "p"
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

	writeTestConfig(t, filepath.Join(tmpDir, "config.dev.yaml"), "server:\n  port: 8080\n")

	t.Setenv("IAM_ENV", "development")
	t.Setenv("GO_ENV", "")
	path := autoConfigFilePath()
	assert.Contains(t, path, "config.dev.yaml")
}

func TestAutoConfigFilePath_ByGoEnv(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(origDir) })

	writeTestConfig(t, filepath.Join(tmpDir, "config.test.yaml"), "server:\n  port: 8080\n")

	t.Setenv("IAM_ENV", "")
	t.Setenv("GO_ENV", "test")
	path := autoConfigFilePath()
	assert.Contains(t, path, "config.test.yaml")
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

func TestValidateConfig_InvalidRedisPort(t *testing.T) {
	cfg := NewConfig()
	cfg.Redis.Port = 99999

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Redis端口无效")
}

func TestLoad_LoadConfigFileErrorStillWorks(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "full.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "test"
jwt:
  secret_key: "env-secret"
mysql:
  password: "env-pass"
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "env-secret", cfg.JWT.SecretKey)
}

func TestLoad_WithRedisPortError(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "bad-redis.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "test"
jwt:
  secret_key: "s"
mysql:
  password: "p"
  port: 3306
redis:
  port: 0
`)

	_, err := Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Redis端口无效")
}
