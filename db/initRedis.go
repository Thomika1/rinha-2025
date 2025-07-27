package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var RedisCtx = context.Background()

func InitRedis() {
	Client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})

}
