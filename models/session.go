package models

import "time"

type Session struct {
	SessionID    string    `bson:"session_id" json:"session_id"`
	UserID       string    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	ClientID     string    `bson:"client_id" json:"client_id"`
	RedirectURI  string    `bson:"redirect_uri" json:"redirect_uri"`
	Scope        string    `bson:"scope" json:"scope"`
	State        string    `bson:"state" json:"state"`
	ResponseType string    `bson:"response_type" json:"response_type"`
	Authenticated bool     `bson:"authenticated" json:"authenticated"`
	ExpiresAt    time.Time `bson:"expires_at" json:"expires_at"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}
