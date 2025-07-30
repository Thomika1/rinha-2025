package main

import (
	"log"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/router"
	"github.com/Thomika1/rinha-2025.git/worker"
	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()

	db.InitRedis()

	worker.InitWorkers()

	router.InitRoutes(app)

	go worker.StartHealthCheckerWithRedis(db.RedisCtx, db.Client, "http://payment-processor-default:8080/payments/service-health", "health:processor:default")
	go worker.StartHealthCheckerWithRedis(db.RedisCtx, db.Client, "http://payment-processor-fallback:8080/payments/service-health", "health:processor:fallback")

	log.Fatal(app.Listen(":8080"))
}
