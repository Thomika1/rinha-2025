package handler

import (
	"encoding/json"
	"log"

	"github.com/Thomika1/rinha-2025.git/db"
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

	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		log.Printf("Erro ao serializar pagamento para JSON: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	err = db.Client.LPush(db.RedisCtx, "payment_jobs", paymentJSON).Err()
	if err != nil {
		log.Printf("Erro ao enfileirar pagamento no Redis: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not queue payment"})
	}

	log.Printf("Pagamento %s enfileirado com sucesso!", payment.CorrelationId)

	// retornar status ok
	return ctx.Status(fiber.StatusAccepted).SendString("Payment queued")
}
