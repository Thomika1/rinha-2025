package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/model"
)

func InitWorkers() {
	log.Print("initializing wokers...")
	processorURLFallback := os.Getenv("PROCESSOR_FALLBACK_URL")
	processorURLDefault := os.Getenv("PROCESSOR_DEFAULT_URL")

	conc := 10

	for i := 0; i < conc; i++ {
		fmt.Printf("\nworker %d initialized", i)
		go func() {
			for {
				result, err := db.Client.BRPop(db.RedisCtx, 0*time.Second, "payment_jobs").Result()
				if err != nil {
					//log.Printf("Error retrieving task from redis :%v waiting...", err)
					//time.Sleep(100 * time.Millisecond)
					continue
				}
				taskData := result[1]
				log.Printf("New task received: %s", taskData)

				var payment model.Payments
				if err := json.Unmarshal([]byte(taskData), &payment); err != nil {
					log.Printf("Error deserializing task: %v", err)
					continue
				}
				paymentJSON, err := json.Marshal(payment)
				if err != nil {
					log.Printf("Error serializing payment")
					continue
				}
				//////////////////////

				defaultStatus, err := getHealthFromRedis(db.RedisCtx, db.Client, "health:processor:default")
				if err != nil {
					//log.Printf("Error getting Default health state from redis Redis: %v. Treating as a failure.", err)
					defaultStatus.Failing = true
				}
				///////////////////////

				if !defaultStatus.Failing {
					//log.Printf("Default processor is healthy. Sending payment...")
					err := PaymentProcessor(payment, processorURLDefault)
					if err != nil {
						//log.Printf("Error processing payment %s", err)
						db.Client.LPush(db.RedisCtx, "payment_jobs", paymentJSON)
						continue
					}
					//updateSummaryCounters(payment, "default")
					continue
				}

				fallbackStatus, err := getHealthFromRedis(db.RedisCtx, db.Client, "health:processor:fallback")
				if err != nil {
					//log.Printf("Error getting Fallback health state from redis Redis: %v. Treating as a failure.", err)
					fallbackStatus.Failing = true
				}

				if !fallbackStatus.Failing {
					//log.Printf("Fallback processor is healthy. Sending payment...")
					err := PaymentProcessor(payment, processorURLFallback)
					if err != nil {
						//log.Printf("Error processing payment %s", err)
						db.Client.LPush(db.RedisCtx, "payment_jobs", paymentJSON)
						continue
					}
					refactorSummary(payment)
					continue
				}

				log.Printf("ALERT: Both processors are failing %s.", payment.CorrelationId)

				db.Client.LPush(db.RedisCtx, "payment_jobs", paymentJSON)
			} // infinite loop, waiting to pop from payments list
		}()
	}
}

func UpdateSummaryCounters(p model.Payments, processorUsed string) error {
	// 1. Converte o valor do pagamento para float64, que é o que IncrByFloat espera.
	// Lembre-se que usamos a biblioteca decimal para precisão.
	amountFloat, _ := p.Amount.Float64()

	// 2. Define as chaves do Redis com base em qual processador foi usado.
	requestsKey := fmt.Sprintf("summary:%s:requests", processorUsed) // ex: "summary:default:requests"
	amountKey := fmt.Sprintf("summary:%s:amount", processorUsed)     // ex: "summary:default:amount"

	// 3. Incrementa o contador de requisições
	if err := db.Client.Incr(db.RedisCtx, requestsKey).Err(); err != nil {
		log.Printf("Erro ao incrementar contador de requisições: %v", err)
		return err
	}

	// 4. Incrementa o valor total
	if err := db.Client.IncrByFloat(db.RedisCtx, amountKey, amountFloat).Err(); err != nil {
		log.Printf("Erro ao incrementar valor total: %v", err)
		return err
	}

	//log.Printf("Contadores de resumo para o processador '%s' atualizados.", processorUsed)
	return nil
}

func refactorSummary(payment model.Payments) error {
	//log.Printf("Refatorando contadores para o CorrelationID: %s. Movendo de 'default' para 'fallback'.", payment.CorrelationId)

	amountFloat, _ := payment.Amount.Float64()

	defaultRequestsKey := "summary:default:requests"
	defaultAmountKey := "summary:default:amount"
	fallbackRequestsKey := "summary:fallback:requests"
	fallbackAmountKey := "summary:fallback:amount"

	pipe := db.Client.Pipeline()

	pipe.Decr(db.RedisCtx, defaultRequestsKey)
	pipe.IncrByFloat(db.RedisCtx, defaultAmountKey, -amountFloat) // Incrementa por um valor negativo para subtrair

	pipe.Incr(db.RedisCtx, fallbackRequestsKey)
	pipe.IncrByFloat(db.RedisCtx, fallbackAmountKey, amountFloat)

	_, err := pipe.Exec(db.RedisCtx)
	if err != nil {
		log.Printf("ERRO CRÍTICO AO REFATORAR CONTADORES: %v", err)
		return fmt.Errorf("falha ao executar pipeline de refatoração de sumário: %w", err)
	}

	//log.Printf("Contadores para o CorrelationID: %s refatorados com sucesso.", payment.CorrelationId)
	return nil
}
