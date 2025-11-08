package utils

import (
	"crypto/rsa"
	"os"
)

// LoadTestKeys loads the test RSA key pair from the keys directory
func LoadTestKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Load private key
	privateKeyPEM, err := os.ReadFile("../keys/private.pem")
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := ParsePrivateKey(string(privateKeyPEM))
	if err != nil {
		return nil, nil, err
	}

	// Load public key
	publicKeyPEM, err := os.ReadFile("../keys/public.pem")
	if err != nil {
		return nil, nil, err
	}

	publicKey, err := ParsePublicKey(string(publicKeyPEM))
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}
