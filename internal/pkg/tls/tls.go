// Package tls 提供服务端 HTTPS 证书加载与 tls.Config 构建的公共工具，
// 供 iam-auth-server / iam-api-server 复用。
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"
)

// LoadServerCert 加载服务端证书与私钥，并主动校验证书有效期，
// 在监听启动前即可暴露证书问题，而非等到握手时才失败。
func LoadServerCert(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("load server cert: %w", err)
	}
	if len(cert.Certificate) == 0 {
		return tls.Certificate{}, fmt.Errorf("load server cert: no certificate found in %s", certFile)
	}

	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("parse server cert: %w", err)
	}
	now := time.Now()
	if now.Before(leaf.NotBefore) {
		return tls.Certificate{}, fmt.Errorf("server cert not yet valid (NotBefore=%s)", leaf.NotBefore.Format(time.RFC3339))
	}
	if now.After(leaf.NotAfter) {
		return tls.Certificate{}, fmt.Errorf("server cert expired (NotAfter=%s)", leaf.NotAfter.Format(time.RFC3339))
	}
	return cert, nil
}

// NewServerTLSConfig 组装服务端 tls.Config。
// 采用静态证书加载（不引入 GetCertificate 动态刷新），最低 TLS1.2。
func NewServerTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := LoadServerCert(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}, nil
}
