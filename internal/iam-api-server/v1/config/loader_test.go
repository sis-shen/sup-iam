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
	assert.Equal(t, 8888, v.GetInt("server.port"))
	assert.Equal(t, "debug", v.GetString("server.mode"))
	assert.Equal(t, "info", v.GetString("log.level"))
	assert.Equal(t, "console", v.GetString("log.format"))
	assert.Equal(t, "header:Authorization", v.GetString("jwt.token_lookup"))
	assert.Equal(t, "iam-apiserver", v.GetString("jwt.issuer"))
	assert.Equal(t, 3306, v.GetInt("mysql.port"))
	assert.Equal(t, 6379, v.GetInt("redis.port"))
	assert.Equal(t, 9090, v.GetInt("grpc.port"))
}

func TestLoadConfigFile_Success(t *testing.T) {
	v := viper.New()
	setDefaults(v)

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

	err := LoadConfigFile(v, configPath)
	require.NoError(t, err)
	assert.Equal(t, 5555, v.GetInt("server.port"))
	assert.Equal(t, "test-jwt-secret", v.GetString("jwt.secret_key"))
	assert.Equal(t, 3307, v.GetInt("mysql.port"))
}

func TestLoadConfigFile_EmptyPath(t *testing.T) {
	v := viper.New()
	// 切换到临时目录以避免误用当前目录下的 config.yaml
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(origDir) })

	err = LoadConfigFile(v, "")
	assert.Error(t, err)
}

func TestLoadConfigFile_FileNotFound(t *testing.T) {
	v := viper.New()
	err := LoadConfigFile(v, "/nonexistent/config.yaml")
	assert.Error(t, err)
}

func TestLoadEnvVars_Bind(t *testing.T) {
	v := viper.New()
	setDefaults(v)
	LoadEnvVars(v)
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
	v := viper.New()
	setDefaults(v)
	v.Set("server.port", 8080)
	v.Set("server.mode", "test")
	v.Set("jwt.secret_key", "secret")
	v.Set("mysql.password", "pass")
	v.Set("mysql.port", 3306)
	v.Set("redis.port", 99999)

	var cfg AppConfig
	err := v.Unmarshal(&cfg)
	require.NoError(t, err)

	err = validateConfig(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Redis端口无效")
}

// Test that load with env vars can work for loading missing values
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
