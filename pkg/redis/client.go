package redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func NewClient(addr string) (*redis.Client, context.Context) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // ← Пустой пароль
		DB:       0,
	})

	ctx := context.Background()

	// Проверка подключения к Redis
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️ Warning: Redis connection failed: %v. Continuing without Redis.", err)
	} else {
		log.Println("✅ Successfully connected to Redis")
	}

	return rdb, ctx
}
