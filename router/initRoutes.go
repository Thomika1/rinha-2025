package router

import (
	"github.com/Thomika1/rinha-2025.git/handler"
	"github.com/gofiber/fiber/v2"
)

func InitRoutes(app *fiber.App) {
	app.Post("/payments", handler.Payments)
	app.Get("/payments-summary", handler.PaymentsSummary)
}
