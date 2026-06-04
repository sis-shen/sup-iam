package pumps

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePrivateKey_PKCS1(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(t, err)

	der := x509.MarshalPKCS1PrivateKey(key)
	parsed, err := parsePrivateKey(der)
	assert.NoError(t, err)
	assert.IsType(t, &rsa.PrivateKey{}, parsed)

	rsaKey := parsed.(*rsa.PrivateKey)
	assert.Equal(t, key.D, rsaKey.D, "Parsed PKCS1 key should have the same D value")
}

func TestParsePrivateKey_PKCS8_RSA(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	assert.NoError(t, err)

	parsed, err := parsePrivateKey(der)
	assert.NoError(t, err)
	assert.IsType(t, &rsa.PrivateKey{}, parsed)
}

func TestParsePrivateKey_PKCS8_ECDSA(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	assert.NoError(t, err)

	parsed, err := parsePrivateKey(der)
	assert.NoError(t, err)
	assert.IsType(t, &ecdsa.PrivateKey{}, parsed)
}

func TestParsePrivateKey_EC(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	der, err := x509.MarshalECPrivateKey(key)
	assert.NoError(t, err)

	parsed, err := parsePrivateKey(der)
	assert.NoError(t, err)
	assert.IsType(t, &ecdsa.PrivateKey{}, parsed)
}

func TestParsePrivateKey_Invalid(t *testing.T) {
	_, err := parsePrivateKey([]byte("invalid key data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestParsePrivateKey_Empty(t *testing.T) {
	_, err := parsePrivateKey([]byte{})
	assert.Error(t, err)
}

func TestParsePrivateKey_Nil(t *testing.T) {
	_, err := parsePrivateKey(nil)
	assert.Error(t, err)
}

func TestParsePrivateKey_PKCS8_UnknownType(t *testing.T) {
	// Test with a valid ASN.1 structure that's not a known key type
	der := []byte{
		0x30, 0x05, // SEQUENCE of length 5
		0x02, 0x01, 0x00, // INTEGER 0
		0x02, 0x00, // INTEGER (empty) - not a valid key
	}
	_, err := parsePrivateKey(der)
	assert.Error(t, err)
}
