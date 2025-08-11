package redisd

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

var (
	Ctx = context.Background()
	RDB *redis.Client
)

func InitRedis() *redis.Client {
	RDB := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "hello world",
		DB:       0,
	})
	if _, err := RDB.Ping(Ctx).Result(); err != nil {

		log.Fatal("i can not connected to redis!", err)
	}

	log.Println("connected to redis")
	return RDB
}
