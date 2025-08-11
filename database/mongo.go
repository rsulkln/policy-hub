package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func initMongo() error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, cErr := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if cErr != nil {

		return cErr
	}
	pErr := mongoClient.Ping(ctx, nil)
	if pErr != nil {

		return pErr
	}
	fmt.Println("mongo connected successfully!")
	Client = mongoClient

	return nil
}
