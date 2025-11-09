package repository

import (
	"context"
	"oauth2-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SSOSessionRepository struct {
	collection *mongo.Collection
}

func NewSSOSessionRepository(db *mongo.Database) *SSOSessionRepository {
	repo := &SSOSessionRepository{
		collection: db.Collection("sso_sessions"),
	}
	
	// Create indexes
	repo.createIndexes(context.Background())
	
	return repo
}

func (r *SSOSessionRepository) createIndexes(ctx context.Context) error {
	// Create unique index on session_id
	sessionIDIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	
	// Create index on user_id
	userIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}
	
	// Create index on expires_at for efficient cleanup
	expiresAtIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "expires_at", Value: 1}},
	}
	
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		sessionIDIndex,
		userIDIndex,
		expiresAtIndex,
	})
	
	return err
}

func (r *SSOSessionRepository) Create(ctx context.Context, session *models.SSOSession) error {
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	if session.LastActivity.IsZero() {
		session.LastActivity = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, session)
	return err
}

func (r *SSOSessionRepository) FindBySessionID(ctx context.Context, sessionID string) (*models.SSOSession, error) {
	var session models.SSOSession
	err := r.collection.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SSOSessionRepository) UpdateLastActivity(ctx context.Context, sessionID string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"session_id": sessionID},
		bson.M{"$set": bson.M{"last_activity": time.Now()}},
	)
	return err
}

func (r *SSOSessionRepository) Delete(ctx context.Context, sessionID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"session_id": sessionID})
	return err
}

func (r *SSOSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.collection.DeleteMany(
		ctx,
		bson.M{"expires_at": bson.M{"$lt": time.Now()}},
	)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (r *SSOSessionRepository) FindByUserID(ctx context.Context, userID string) ([]*models.SSOSession, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var sessions []*models.SSOSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}
	
	return sessions, nil
}
