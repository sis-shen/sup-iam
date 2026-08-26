package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConfig_TLSDefaults 验证默认配置：server.tls 默认关闭，向后兼容。
func TestNewConfig_TLSDefaults(t *testing.T) {
	cfg := NewConfig()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Server)

	assert.False(t, cfg.Server.TLS.Enabled)
	assert.Empty(t, cfg.Server.TLS.CertFile)
	assert.Empty(t, cfg.Server.TLS.KeyFile)
}

// TestLoad_ServerTLSEnabled 验证配置文件中 server.tls.enabled=true 能被正确解析，
// 且不影响既有 validateConfig 校验。
func TestLoad_ServerTLSEnabled(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "test"
  tls:
    enabled: true
    cert_file: "/etc/iam/tls/server.crt"
    key_file: "/etc/iam/tls/server.key"
jwt:
  secret_key: "test-jwt-secret"
mysql:
  password: "test-pass"
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.True(t, cfg.Server.TLS.Enabled)
	assert.Equal(t, "/etc/iam/tls/server.crt", cfg.Server.TLS.CertFile)
	assert.Equal(t, "/etc/iam/tls/server.key", cfg.Server.TLS.KeyFile)
}

// TestLoad_ServerTLSDisabled 验证显式配置 enabled=false 时保持关闭。
func TestLoad_ServerTLSDisabled(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "test"
  tls:
    enabled: false
jwt:
  secret_key: "test-jwt-secret"
mysql:
  password: "test-pass"
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.False(t, cfg.Server.TLS.Enabled)
	assert.Empty(t, cfg.Server.TLS.CertFile)
	assert.Empty(t, cfg.Server.TLS.KeyFile)
}

// TestLoad_ServerTLSDefaultDisabled 验证未配置 server.tls 段时默认为 enabled=false（向后兼容）。
func TestLoad_ServerTLSDefaultDisabled(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	writeTestConfig(t, configPath, `
server:
  port: 8080
  mode: "test"
jwt:
  secret_key: "test-jwt-secret"
mysql:
  password: "test-pass"
`)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.False(t, cfg.Server.TLS.Enabled)
	assert.Empty(t, cfg.Server.TLS.CertFile)
	assert.Empty(t, cfg.Server.TLS.KeyFile)
}
