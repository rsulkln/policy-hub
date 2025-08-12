package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"project/database/mongo"
	redisd "project/database/redis"
	"project/model"
	"time"
)

type MongoUserRepository struct {
	collection *mongoDriver.Collection
}

func NewMongoUserRepository(client *mongoDriver.Client, dbName, CollectionName string) *MongoUserRepository {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := client.Database(dbName).Collection(CollectionName)
	return &MongoUserRepository{collection: collection}
}

func (repo *MongoUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := repo.collection.InsertOne(ctx, user); err != nil {
		return err
	}
	return nil
}
func (repo *MongoUserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {

	var user model.User
	gErr := redisd.GetCash(id, user)
	if gErr != nil {
		return nil, gErr
	}
	if user.ID == "" {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	err := repo.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (repo *MongoUserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	var user model.User
	if err := mongo.DB.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		return nil, err

	} else {
		return &user, err
	}

}
