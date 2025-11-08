package main

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"
	"oauth2-server/config"
	"oauth2-server/database"
	"oauth2-server/handlers"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.Load()

	privateKey, publicKey, err := loadOrGenerateKeys()
	if err != nil {
		log.Fatalf("Failed to load keys: %v", err)
	}
	cfg.PrivateKey = privateKey
	cfg.PublicKey = publicKey

	db, err := database.Connect(cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to MongoDB successfully")

	if err := createIndexes(db.DB); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	userRepo := repository.NewUserRepository(db.DB)
	clientRepo := repository.NewClientRepository(db.DB)
	authCodeRepo := repository.NewAuthCodeRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)

	authHandler := handlers.NewAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, cfg)
	oauthHandler := handlers.NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, cfg)
	clientHandler := handlers.NewClientHandler(clientRepo)
	discoveryHandler := handlers.NewDiscoveryHandler("http://localhost:" + cfg.ServerPort)
	jwksHandler := handlers.NewJWKSHandler(publicKey)
	tokenExchangeHandler := handlers.NewTokenExchangeHandler(userRepo, clientRepo, cfg)
	tokenValidationHandler := handlers.NewTokenValidationHandler(cfg)

	r := mux.NewRouter()

	r.HandleFunc("/.well-known/openid-configuration", discoveryHandler.WellKnown).Methods("GET")
	r.HandleFunc("/.well-known/jwks.json", jwksHandler.JWKS).Methods("GET")

	r.HandleFunc("/auth/register", authHandler.ShowRegister).Methods("GET")
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.ShowLogin).Methods("GET")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	r.HandleFunc("/oauth/authorize", oauthHandler.Authorize).Methods("GET")
	r.HandleFunc("/oauth/token", oauthHandler.Token).Methods("POST")
	r.HandleFunc("/oauth/userinfo", oauthHandler.UserInfo).Methods("GET")

	r.HandleFunc("/token/exchange", tokenExchangeHandler.HandleTokenExchange).Methods("POST")
	r.HandleFunc("/token/validate", tokenValidationHandler.ValidateToken).Methods("GET", "POST")

	r.HandleFunc("/clients/register", clientHandler.RegisterClient).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	log.Printf("OAuth2 Server starting on port %s", cfg.ServerPort)
	log.Printf("Using RS256 for JWT signing")
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, r))
}

func loadOrGenerateKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKeyPath := "keys/private.pem"
	publicKeyPath := "keys/public.pem"

	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		log.Println("Generating new RSA key pair...")

		if err := os.MkdirAll("keys", 0700); err != nil {
			return nil, nil, err
		}

		privateKey, err := utils.GenerateRSAKeyPair(2048)
		if err != nil {
			return nil, nil, err
		}

		if err := utils.SavePrivateKeyToFile(privateKey, privateKeyPath); err != nil {
			return nil, nil, err
		}

		if err := utils.SavePublicKeyToFile(&privateKey.PublicKey, publicKeyPath); err != nil {
			return nil, nil, err
		}

		log.Println("RSA key pair generated and saved")
		return privateKey, &privateKey.PublicKey, nil
	}

	log.Println("Loading existing RSA key pair...")
	privateKey, err := utils.LoadPrivateKeyFromFile(privateKeyPath)
	if err != nil {
		return nil, nil, err
	}

	publicKey, err := utils.LoadPublicKeyFromFile(publicKeyPath)
	if err != nil {
		return nil, nil, err
	}

	log.Println("RSA key pair loaded successfully")
	return privateKey, publicKey, nil
}

func createIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersCollection := db.Collection("users")
	_, err := usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	clientsCollection := db.Collection("clients")
	_, err = clientsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "client_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	authCodesCollection := db.Collection("auth_codes")
	_, err = authCodesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = authCodesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		return err
	}

	sessionsCollection := db.Collection("sessions")
	_, err = sessionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = sessionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		return err
	}

	log.Println("Database indexes created successfully")
	return nil
}
