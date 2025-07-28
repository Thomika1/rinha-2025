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

	router.InitRoutes(app)

	go worker.InitWorkers()

	log.Fatal(app.Listen(":8080"))
}
