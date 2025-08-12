package redisd

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var (
	Ctx = context.Background()
	RDB *redis.Client
)

func InitRedis() *redis.Client {
	RDB = redis.NewClient(&redis.Options{
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

func SetCash(key string, value interface{}, ttl time.Duration) error {
	jsonData, mErr := json.Marshal(value)
	if mErr != nil {
		return mErr
	}
	return RDB.Set(Ctx, key, jsonData, ttl).Err()
}

func GetCash(key string, dest interface{}) error {
	val, gErr := RDB.Get(Ctx, key).Result()
	if gErr == redis.Nil {
		return nil
	}
	if gErr != nil {
		return gErr
	}
	return json.Unmarshal([]byte(val), dest)
}

func DelCash(key string) error {
	return RDB.Del(Ctx, key).Err()
}
