package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var RedisCtx = context.Background()

func InitRedis() {
	Client = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})

	_, err := Client.Ping(RedisCtx).Result()
	if err != nil {
		// Use log.Fatalf para encerrar a aplicação se a conexão falhar.
		log.Fatalf("Não foi possível conectar ao Redis: %v", err)
	}

	log.Println("Conectado ao Redis com sucesso!")
}
