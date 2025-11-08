package config

import (
	"crypto/rsa"
	"os"
	"strconv"
)

type Config struct {
	MongoURI            string
	DatabaseName        string
	PrivateKey          *rsa.PrivateKey
	PublicKey           *rsa.PublicKey
	ServerPort          string
	AccessTokenExpiry   int64
	RefreshTokenExpiry  int64
}

func Load() *Config {
	return &Config{
		MongoURI:            getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DatabaseName:        getEnv("DATABASE_NAME", "oauth2_db"),
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		AccessTokenExpiry:   getEnvAsInt("ACCESS_TOKEN_EXPIRY", 3600),
		RefreshTokenExpiry:  getEnvAsInt("REFRESH_TOKEN_EXPIRY", 604800),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}
