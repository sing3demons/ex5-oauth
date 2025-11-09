package repository

import (
	"context"
	"oauth2-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserConsentRepository struct {
	collection *mongo.Collection
}

func NewUserConsentRepository(db *mongo.Database) *UserConsentRepository {
	repo := &UserConsentRepository{
		collection: db.Collection("user_consents"),
	}
	
	// Create indexes
	repo.createIndexes(context.Background())
	
	return repo
}

func (r *UserConsentRepository) createIndexes(ctx context.Context) error {
	// Create unique compound index on user_id + client_id
	userClientIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "client_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	
	// Create index on user_id for listing user consents
	userIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}
	
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		userClientIndex,
		userIDIndex,
	})
	
	return err
}

func (r *UserConsentRepository) Create(ctx context.Context, consent *models.UserConsent) error {
	if consent.GrantedAt.IsZero() {
		consent.GrantedAt = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, consent)
	return err
}

func (r *UserConsentRepository) FindByUserAndClient(ctx context.Context, userID, clientID string) (*models.UserConsent, error) {
	var consent models.UserConsent
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":   userID,
		"client_id": clientID,
	}).Decode(&consent)
	if err != nil {
		return nil, err
	}
	return &consent, nil
}

func (r *UserConsentRepository) HasConsent(ctx context.Context, userID, clientID string, scopes []string) (bool, error) {
	consent, err := r.FindByUserAndClient(ctx, userID, clientID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	
	// Check if consent is expired
	if !consent.ExpiresAt.IsZero() && consent.ExpiresAt.Before(time.Now()) {
		return false, nil
	}
	
	// Check if all requested scopes are included in the stored consent
	consentScopeMap := make(map[string]bool)
	for _, scope := range consent.Scopes {
		consentScopeMap[scope] = true
	}
	
	for _, requestedScope := range scopes {
		if !consentScopeMap[requestedScope] {
			return false, nil
		}
	}
	
	return true, nil
}

func (r *UserConsentRepository) RevokeConsent(ctx context.Context, userID, clientID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{
		"user_id":   userID,
		"client_id": clientID,
	})
	return err
}

func (r *UserConsentRepository) ListUserConsents(ctx context.Context, userID string) ([]*models.UserConsent, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var consents []*models.UserConsent
	if err := cursor.All(ctx, &consents); err != nil {
		return nil, err
	}
	
	return consents, nil
}
