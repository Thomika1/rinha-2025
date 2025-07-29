package worker

import (
	"encoding/json"
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

	for {

		result, err := db.Client.BRPop(db.RedisCtx, 0*time.Second, "payment_jobs").Result()
		if err != nil {
			log.Printf("Error retrieving task from redis :%v waiting...", err)
			time.Sleep(1 * time.Second)
			continue
		}

		taskData := result[1]
		log.Printf("New task received: %s", taskData)

		var payment model.Payments
		if err := json.Unmarshal([]byte(taskData), &payment); err != nil {
			log.Printf("Error deserializing task: %v", err)
			continue
		}

		// 1. Tenta verificar a saúde do processador Default
		defaultStatus, err := getHealthFromRedis(db.RedisCtx, db.Client, "health:processor:default")
		if err != nil {
			log.Printf("Erro ao obter saúde do processador Default do Redis: %v. Tratando como falha.", err)
			defaultStatus.Failing = true // Trata erro de conexão com Redis como falha do serviço
		}

		// 2. Se o Default estiver OK, processe o pagamento e pule para a próxima tarefa do loop
		if !defaultStatus.Failing {
			log.Printf("Processador Default está saudável. Enviando pagamento para ele.")

			err := PaymentProcessor(payment, processorURLDefault)
			if err != nil {
				log.Printf("Error processing payment %s", err)
			}
			continue // Tarefa concluída, volta para o início do loop `for`
		}

		// 3. Se o Default falhou, tenta verificar a saúde do processador Fallback
		log.Println("Processador Default em falha. Verificando o Fallback...")
		fallbackStatus, err := getHealthFromRedis(db.RedisCtx, db.Client, "health:processor:fallback")
		if err != nil {
			log.Printf("Erro ao obter saúde do processador Fallback do Redis: %v. Tratando como falha.", err)
			fallbackStatus.Failing = true
		}

		// 4. Se o Fallback estiver OK, processe o pagamento
		if !fallbackStatus.Failing {
			log.Printf("Processador Fallback está saudável. Enviando pagamento para ele.")

			err := PaymentProcessor(payment, processorURLFallback)
			if err != nil {
				log.Printf("Error processing payment %s", err)
			}
			continue // Tarefa concluída
		}

		// 5. Se AMBOS falharam, você precisa de uma estratégia de falha
		log.Printf("ALERTA: Ambos os processadores de pagamento estão em falha para a tarefa %s.", payment.CorrelationId)

		paymentJSON, err := json.Marshal(payment)
		if err != nil {
			log.Printf("Error serializing payment for requeueing")
		}
		db.Client.LPush(db.RedisCtx, "payment_jobs", paymentJSON)

	} // infinite loop, waiting to pop from payments list

}
