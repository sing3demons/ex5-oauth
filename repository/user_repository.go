package repository

import (
	"context"
	"oauth2-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	// Set the ID from the inserted document
	if oid, ok := result.InsertedID.(string); ok {
		user.ID = oid
	} else if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid.Hex()
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	// Try to find by string ID first
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == nil {
		return &user, nil
	}
	
	// If not found, try as ObjectID
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
		if err == nil {
			return &user, nil
		}
	}
	
	return nil, err
}
