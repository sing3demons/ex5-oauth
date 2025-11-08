package handlers

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"net/http"
)

type JWKSHandler struct {
	publicKey *rsa.PublicKey
}

func NewJWKSHandler(publicKey *rsa.PublicKey) *JWKSHandler {
	return &JWKSHandler{
		publicKey: publicKey,
	}
}

func (h *JWKSHandler) JWKS(w http.ResponseWriter, r *http.Request) {
	n := base64.RawURLEncoding.EncodeToString(h.publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(h.publicKey.E)).Bytes())

	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": "1",
				"n":   n,
				"e":   e,
			},
		},
	}

	respondJSON(w, http.StatusOK, jwks)
}
