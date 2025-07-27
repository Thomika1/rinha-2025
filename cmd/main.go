package main

import (
	"log"

	"github.com/Thomika1/rinha-2025.git/handler"
	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()

	handler.InitRoutes(app)

	log.Fatal(app.Listen(":8080"))
}
