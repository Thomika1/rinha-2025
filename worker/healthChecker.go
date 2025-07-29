package worker

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Thomika1/rinha-2025.git/model"
	"github.com/redis/go-redis/v9"
)

func StartHealthCheckerWithRedis(ctx context.Context, redisClient *redis.Client, url string, redisKey string) {
	check := func() {
		resp, err := http.Get(url)
		var newStatus model.ServiceHealth

		if err != nil {
			log.Printf("Error checking health state %s: %v", url, err)
			newStatus.Failing = true
		} else {
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&newStatus); err != nil {
				log.Printf("Error decoding health state from %s: %v", url, err)
				newStatus.Failing = true
			}
		}

		// --- LÓGICA DO REDIS ---
		// 1. Serializa o status para JSON
		statusJSON, err := json.Marshal(newStatus)
		if err != nil {
			log.Printf("Error serializing health state to json: %v", err)
			return
		}

		// 2. Salva o JSON como uma string no Redis. Usamos SET (sobrescreve o valor anterior).
		// Não definimos expiração (0), pois o worker atualiza o valor a cada 5s.
		if err := redisClient.Set(ctx, redisKey, statusJSON, 0).Err(); err != nil {
			log.Printf("Error seting health sate on redis %s: %v", redisKey, err)
		}
		// ----------------------

		log.Printf("Health state for %s was updated", redisKey)
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

	// Pega o JSON do Redis
	statusJSON, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		// Chave não existe ainda, talvez o 1º health check não rodou.
		// Assumimos como falha para segurança.
		return model.ServiceHealth{Failing: true}, nil
	} else if err != nil {
		return model.ServiceHealth{}, err // Erro de conexão com o Redis
	}

	// Desserializa o JSON para o struct
	if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
		return model.ServiceHealth{}, err
	}

	return status, nil
}
