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
			time.Sleep(100 * time.Millisecond)
			continue
		}

		taskData := result[1]
		log.Printf("New task received: %s", taskData)

		var payment model.Payments
		if err := json.Unmarshal([]byte(taskData), &payment); err != nil {
			log.Printf("Error deserializing task: %v", err)
			continue
		}

		defaultStatus, err := getHealthFromRedis(db.RedisCtx, db.Client, "health:processor:default")
		if err != nil {
			log.Printf("Error getting Default health state from redis Redis: %v. Treating as a failure.", err)
			defaultStatus.Failing = true
		}

		if !defaultStatus.Failing {
			log.Printf("Default processor is healthy. Sending payment...")

			err := PaymentProcessor(payment, processorURLDefault)
			if err != nil {
				log.Printf("Error processing payment %s", err)
			}
			continue
		}

		log.Println("Default processor failing. Verifying Fallback...")
		fallbackStatus, err := getHealthFromRedis(db.RedisCtx, db.Client, "health:processor:fallback")
		if err != nil {
			log.Printf("Error getting Fallback health state from redis Redis: %v. Treating as a failure.", err)
			fallbackStatus.Failing = true
		}

		if !fallbackStatus.Failing {
			log.Printf("Fallback processor is healthy. Sending payment...")

			err := PaymentProcessor(payment, processorURLFallback)
			if err != nil {
				log.Printf("Error processing payment %s", err)
			}
			continue
		}

		log.Printf("ALERT: Both processors are failing %s.", payment.CorrelationId)

		paymentJSON, err := json.Marshal(payment)
		if err != nil {
			log.Printf("Error serializing payment for requeueing")
		}
		db.Client.LPush(db.RedisCtx, "payment_jobs", paymentJSON)

	} // infinite loop, waiting to pop from payments list

}
