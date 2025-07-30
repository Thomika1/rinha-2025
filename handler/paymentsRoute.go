package handler

import (
	"strconv"

	"github.com/bytedance/sonic"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/model"
	"github.com/Thomika1/rinha-2025.git/worker"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

func Payments(ctx *fiber.Ctx) error {

	payment := new(model.Payments)

	if err := ctx.BodyParser(payment); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "could not parse JSON",
		})
	}
	// enfileirar o payment na fila do redis

	paymentJSON, err := sonic.Marshal(payment)
	if err != nil {
		//log.Printf("Error serializing payment to JSON: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	err = db.Client.LPush(db.RedisCtx, "payment_jobs", paymentJSON).Err()
	if err != nil {
		//log.Printf("Error queuing payment: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not queue payment"})
	}

	worker.UpdateSummaryCounters(*payment, "default")

	//log.Printf("\nPayment %s succssefuly queued!", payment.CorrelationId)

	// retornar status ok
	return ctx.Status(fiber.StatusAccepted).SendString("Payment queued")
}

func PaymentsSummary(ctx *fiber.Ctx) error {

	var response model.PaymentsSummaryResponse

	keys := []string{
		"summary:default:requests",
		"summary:default:amount",
		"summary:fallback:requests",
		"summary:fallback:amount",
	}

	results, err := db.Client.MGet(db.RedisCtx, keys...).Result()
	if err != nil {
		//log.Printf("failed to retrieve summary from redis: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve summary from redis",
		})
	}
	if results[0] != nil {
		response.Default.TotalRequests, _ = strconv.ParseInt(results[0].(string), 10, 64)
	}
	if results[1] != nil {
		request, _ := decimal.NewFromString(results[1].(string))
		response.Default.TotalAmount = request
	}
	if results[2] != nil {
		request, _ := strconv.ParseInt(results[2].(string), 10, 64)
		response.Fallback.TotalRequests = request
	}
	if results[3] != nil {
		response.Fallback.TotalAmount, _ = decimal.NewFromString(results[3].(string))
	}

	// 5. Retorna a resposta final em formato JSON.
	return ctx.Status(fiber.StatusOK).JSON(response)
}
