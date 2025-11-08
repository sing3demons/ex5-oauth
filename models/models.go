package models

import (
	"time"
)

type User struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	Email        string    `bson:"email" json:"email"`
	Password     string    `bson:"password" json:"-"`
	Name         string    `bson:"name" json:"name"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}

type Client struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	ClientID     string    `bson:"client_id" json:"client_id"`
	ClientSecret string    `bson:"client_secret" json:"-"`
	RedirectURIs []string  `bson:"redirect_uris" json:"redirect_uris"`
	Name         string    `bson:"name" json:"name"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}

type AuthorizationCode struct {
	Code        string    `bson:"code" json:"code"`
	ClientID    string    `bson:"client_id" json:"client_id"`
	UserID      string    `bson:"user_id" json:"user_id"`
	RedirectURI string    `bson:"redirect_uri" json:"redirect_uri"`
	Scope       string    `bson:"scope" json:"scope"`
	ExpiresAt   time.Time `bson:"expires_at" json:"expires_at"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type UserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
