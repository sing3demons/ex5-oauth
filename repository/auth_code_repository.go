package repository

import (
	"context"
	"oauth2-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthCodeRepository struct {
	collection *mongo.Collection
}

func NewAuthCodeRepository(db *mongo.Database) *AuthCodeRepository {
	return &AuthCodeRepository{
		collection: db.Collection("auth_codes"),
	}
}

func (r *AuthCodeRepository) Create(ctx context.Context, code *models.AuthorizationCode) error {
	code.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, code)
	return err
}

func (r *AuthCodeRepository) FindByCode(ctx context.Context, code string) (*models.AuthorizationCode, error) {
	var authCode models.AuthorizationCode
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&authCode)
	if err != nil {
		return nil, err
	}
	return &authCode, nil
}

func (r *AuthCodeRepository) Delete(ctx context.Context, code string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"code": code})
	return err
}
