package worker

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/model"
)

func InitWorkers() {
	log.Print("initializing wokers...")

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

	} // infinite loop, waiting to pop from payments list

}
