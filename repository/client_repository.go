package repository

import (
	"context"
	"oauth2-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ClientRepository struct {
	collection *mongo.Collection
}

func NewClientRepository(db *mongo.Database) *ClientRepository {
	return &ClientRepository{
		collection: db.Collection("clients"),
	}
}

func (r *ClientRepository) Create(ctx context.Context, client *models.Client) error {
	client.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, client)
	return err
}

func (r *ClientRepository) FindByClientID(ctx context.Context, clientID string) (*models.Client, error) {
	var client models.Client
	err := r.collection.FindOne(ctx, bson.M{"client_id": clientID}).Decode(&client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
