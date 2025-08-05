package worker

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bytedance/sonic"

	"github.com/Thomika1/rinha-2025.git/model"
	"github.com/redis/go-redis/v9"
)

func StartHealthCheckerWithRedis(ctx context.Context, redisClient *redis.Client, url string, redisKey string) {
	check := func() {
		resp, err := http.Get(url)
		var newStatus model.ServiceHealth

		if err != nil {
			newStatus.Failing = true
		} else {
			defer resp.Body.Close()

			// 1. Lê todo o corpo da resposta para um slice de bytes.
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				newStatus.Failing = true
			} else {
				// 2. Usa sonic.Unmarshal para decodificar o slice de bytes.
				if err := sonic.Unmarshal(body, &newStatus); err != nil {
					newStatus.Failing = true
				}
			}
		}

		statusJSON, err := sonic.Marshal(newStatus)
		if err != nil {
			log.Printf("Error serializing health state to json: %v", err)
			return
		}
		ctx = context.Background()
		if err := redisClient.Set(ctx, redisKey, statusJSON, 0).Err(); err != nil {
			log.Printf("Error seting health sate on redis %s: %v", redisKey, err)
		}

		//log.Printf("Health state for %s was updated", redisKey)
	}

	check() // Verificação imediata

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		check()
	}
}

func getHealthFromRedis(ctx context.Context, redisClient *redis.Client, key string) (model.ServiceHealth, error) {
	var status model.ServiceHealth

	statusJSON, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return model.ServiceHealth{Failing: true}, nil
	} else if err != nil {
		return model.ServiceHealth{}, err // Erro de conexão com o Redis
	}

	// Desserializa o JSON para o struct
	if err := sonic.Unmarshal([]byte(statusJSON), &status); err != nil {
		return model.ServiceHealth{}, err
	}

	return status, nil
}
