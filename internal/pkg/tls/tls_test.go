package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// genSelfSigned 生成一张自签证书并写入临时目录，返回 cert/key 文件路径。
// notAfter 控制证书有效期，用于测试「过期证书」用例。
func genSelfSigned(t *testing.T, notAfter time.Time) (certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "iam-self-signed-test"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	dir := t.TempDir()
	certFile = filepath.Join(dir, "tls.crt")
	keyFile = filepath.Join(dir, "tls.key")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	require.NoError(t, os.WriteFile(certFile, certPEM, 0o600))

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	require.NoError(t, os.WriteFile(keyFile, keyPEM, 0o600))

	return certFile, keyFile
}

func TestLoadServerCert_Valid(t *testing.T) {
	certFile, keyFile := genSelfSigned(t, time.Now().Add(24*time.Hour))

	cert, err := LoadServerCert(certFile, keyFile)
	require.NoError(t, err)
	require.Len(t, cert.Certificate, 1)
}

func TestLoadServerCert_MissingFiles(t *testing.T) {
	_, err := LoadServerCert(filepath.Join(t.TempDir(), "nonexist.crt"), filepath.Join(t.TempDir(), "nonexist.key"))
	require.Error(t, err)
}

func TestLoadServerCert_Expired(t *testing.T) {
	certFile, keyFile := genSelfSigned(t, time.Now().Add(-time.Hour))

	_, err := LoadServerCert(certFile, keyFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestLoadServerCert_InvalidContent(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "tls.crt")
	keyFile := filepath.Join(dir, "tls.key")
	require.NoError(t, os.WriteFile(certFile, []byte("not a pem"), 0o600))
	require.NoError(t, os.WriteFile(keyFile, []byte("not a pem"), 0o600))

	_, err := LoadServerCert(certFile, keyFile)
	require.Error(t, err)
}

func TestNewServerTLSConfig(t *testing.T) {
	certFile, keyFile := genSelfSigned(t, time.Now().Add(24*time.Hour))

	cfg, err := NewServerTLSConfig(certFile, keyFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, uint16(0x0303), cfg.MinVersion, "MinVersion 应为 TLS1.2")
	require.Len(t, cfg.Certificates, 1)
}
