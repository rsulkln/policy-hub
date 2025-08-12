package redisd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
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
		log.Fatal("I can't connect to redis!", err)
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

func SetDualKeyCash(id int, username string, value interface{}, ttl time.Duration) error {

	mainKey := fmt.Sprintf("data:%d", id)

	refKey := fmt.Sprintf("ref:%s", username)

	jsonData, mErr := json.Marshal(value)
	if mErr != nil {
		return mErr
	}

	pipe := RDB.Pipeline()

	pipe.Set(Ctx, mainKey, jsonData, ttl)

	pipe.Set(Ctx, refKey, id, ttl)

	_, err := pipe.Exec(Ctx)
	return err
}

func GetCashByID(id int, dest interface{}) error {
	key := fmt.Sprintf("data:%d", id)
	return GetCash(key, dest)
}

func GetCashByKey(username string, dest interface{}) error {

	refKey := fmt.Sprintf("ref:%s", username)

	val, gErr := RDB.Get(Ctx, refKey).Result()
	if gErr == redis.Nil {
		return nil
	}
	if gErr != nil {
		return gErr
	}

	id, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("invalid ID in reference: %v", err)
	}

	return GetCashByID(id, dest)
}

func UpdateDualKeyCash(id int, value interface{}, ttl time.Duration) error {
	mainKey := fmt.Sprintf("data:%d", id)
	return SetCash(mainKey, value, ttl)
}

func DelDualKeyCash(id int, uniqueKey string) error {
	mainKey := fmt.Sprintf("data:%d", id)
	refKey := fmt.Sprintf("ref:%s", uniqueKey)

	return RDB.Del(Ctx, mainKey, refKey).Err()
}

func DelCashByID(id int) error {
	mainKey := fmt.Sprintf("data:%d", id)

	// اول ببین دیتا وجود داره
	val, err := RDB.Get(Ctx, mainKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil // already deleted
		}
		return err
	}

	// سعی کن unique key رو از دیتا extract کنی
	var temp map[string]interface{}
	if json.Unmarshal([]byte(val), &temp) == nil {
		if keyVal, exists := temp["key"]; exists {
			if keyStr, ok := keyVal.(string); ok {
				refKey := fmt.Sprintf("ref:%s", keyStr)
				return RDB.Del(Ctx, mainKey, refKey).Err()
			}
		}
	}

	return DelCash(mainKey)
}

func DelCashByKey(uniqueKey string) error {
	refKey := fmt.Sprintf("ref:%s", uniqueKey)

	val, err := RDB.Get(Ctx, refKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	id, err := strconv.Atoi(val)
	if err != nil {
		return DelCash(refKey)
	}

	mainKey := fmt.Sprintf("data:%d", id)
	return RDB.Del(Ctx, mainKey, refKey).Err()
}
