package router

import (
	"github.com/Thomika1/rinha-2025.git/handler"
	"github.com/gofiber/fiber/v2"
)

func InitRoutes(app *fiber.App) {
	app.Get("/helloworld", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/payments", handler.Payments)
}
