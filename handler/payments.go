package handler

import (
	"github.com/Thomika1/rinha-2025.git/model"
	"github.com/gofiber/fiber/v2"
)

func Payments(ctx *fiber.Ctx) error {

	payment := new(model.Payments)

	if err := ctx.BodyParser(payment); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}
	// enfileirar o payment na fila do redis

	// retornar status ok
}
