package repository

import (
	"context"
	"oauth2-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SessionRepository struct {
	collection *mongo.Collection
}

func NewSessionRepository(db *mongo.Database) *SessionRepository {
	return &SessionRepository{
		collection: db.Collection("sessions"),
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	session.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, session)
	return err
}

func (r *SessionRepository) FindBySessionID(ctx context.Context, sessionID string) (*models.Session, error) {
	var session models.Session
	err := r.collection.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) Update(ctx context.Context, session *models.Session) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"session_id": session.SessionID},
		bson.M{"$set": session},
	)
	return err
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"session_id": sessionID})
	return err
}
