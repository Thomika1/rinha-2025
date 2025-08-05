package main

import (
	"log"
	"os"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/router"
	"github.com/Thomika1/rinha-2025.git/worker"
	"github.com/gofiber/fiber/v2"
)

func main() {

	db.InitRedis()

	app := fiber.New()
	router.InitRoutes(app)
	worker.InitWorkers()

	api := os.Getenv("API_NAME")
	if api == "1" {
		go worker.StartHealthCheckerWithRedis(db.RedisCtx, db.Client, "http://payment-processor-default:8080/payments/service-health", "health:processor:default")
		go worker.StartHealthCheckerWithRedis(db.RedisCtx, db.Client, "http://payment-processor-fallback:8080/payments/service-health", "health:processor:fallback")
	}

	log.Fatal(app.Listen(":8080"))
}
