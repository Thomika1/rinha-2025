package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bytedance/sonic"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/model"
)

func InitWorkers() {
	log.Print("initializing wokers...")

	conc := 16

	for i := 0; i < conc; i++ {
		fmt.Printf("\nworker %d initialized", i)
		go func() {
			for {
				result, err := db.Client.BRPop(context.Background(), 0, "payment_jobs").Result()
				if err != nil {
					//log.Printf("WORKER ERRO ")
					time.Sleep(1 * time.Second)
					continue
				}
				//fmt.Println("WORKER")
				//log.Printf("New task received: %s", taskData)

				var payment model.Payments
				if err := sonic.Unmarshal([]byte(result[1]), &payment); err != nil {
					log.Printf("Error deserializing task: %v", err)
					continue
				}
				//fmt.Printf("WORKER %s", payment)
				err = PaymentProcessor(payment)
				if err != nil {
					log.Printf("Error processing payment %s", err)

					err := db.Client.RPush(db.RedisCtx, "payment_jobs", payment)

					if err != nil {
						//log.Printf("WORKER Error queuing payment: %v", err)
					}
				}
				// time.Sleep(500 * time.Millisecond)
				continue
			} // infinite loop, waiting to pop from payments list
		}()
	}
}
