package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"sub"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Scope  string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type IDTokenClaims struct {
	jwt.RegisteredClaims
	// Additional claims are added dynamically via MapClaims
}

type AccessTokenClaims struct {
	UserID   string `json:"sub"`
	Scope    string `json:"scope"`
	ClientID string `json:"client_id,omitempty"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID, email, name, scope string, privateKey *rsa.PrivateKey, expiry int64) (string, error) {
	claims := AccessTokenClaims{
		UserID: userID,
		Scope:  scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

type RefreshTokenClaims struct {
	UserID string `json:"sub"`
	Scope  string `json:"scope"`
	jwt.RegisteredClaims
}

func GenerateRefreshToken(userID, scope string, privateKey *rsa.PrivateKey, expiry int64) (string, error) {
	claims := RefreshTokenClaims{
		UserID: userID,
		Scope:  scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// GenerateIDToken generates an ID token with filtered claims based on scopes
func GenerateIDToken(userID, clientID string, userClaims map[string]interface{}, privateKey *rsa.PrivateKey, expiry int64) (string, error) {
	// Start with user claims (already filtered by scope)
	claims := jwt.MapClaims{}
	
	// Add all user claims
	for key, value := range userClaims {
		claims[key] = value
	}
	
	// Add standard JWT claims
	claims["sub"] = userID
	claims["aud"] = clientID
	claims["exp"] = time.Now().Add(time.Duration(expiry) * time.Second).Unix()
	claims["iat"] = time.Now().Unix()
	claims["iss"] = "oauth2-server"

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// GenerateIDTokenLegacy generates an ID token with explicit claims (deprecated, use GenerateIDToken with filtered claims)
func GenerateIDTokenLegacy(userID, email, name, clientID string, privateKey *rsa.PrivateKey, expiry int64) (string, error) {
	claims := jwt.MapClaims{
		"sub":            userID,
		"email":          email,
		"email_verified": true,
		"name":           name,
		"aud":            clientID,
		"exp":            time.Now().Add(time.Duration(expiry) * time.Second).Unix(),
		"iat":            time.Now().Unix(),
		"iss":            "oauth2-server",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func ValidateToken(tokenString string, publicKey *rsa.PublicKey) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func ValidateRefreshToken(tokenString string, publicKey *rsa.PublicKey) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func ParsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaKey, nil
}

func ParsePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaKey, nil
}
