package database

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"project/model"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(client *mongo.Client, dbName, CollectionName string) *MongoUserRepository {
	collection := client.Database(dbName).Collection(CollectionName)
	return &MongoUserRepository{collection: collection}
}

func (repo *MongoUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	if _, err := repo.collection.InsertOne(ctx, user); err != nil {
		return err
	}
	return nil
}
func (repo *MongoUserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	filter := bson.M{"_id": id}
	err := repo.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
