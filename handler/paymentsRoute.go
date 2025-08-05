package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/shopspring/decimal"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/model"
	"github.com/gofiber/fiber/v2"
)

func Payments(ctx *fiber.Ctx) error {
	if ctx.Method() != fiber.MethodPost {
		return ctx.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"error": "Method not allowed",
		})
	}

	var payment model.Payments

	if err := ctx.BodyParser(&payment); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "could not parse JSON",
		})
	}

	//fmt.Printf("HANDLER pre marshal%s", payment)

	paymentJSON, err := sonic.Marshal(payment)
	//fmt.Printf("\nHANDLER pos marshal%s", payment)
	if err != nil {
		log.Printf("Error serializing payment to JSON: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	//fmt.Printf("HANDLER struct: %+v\n", payment)
	//fmt.Printf("HANDLER JSON: %s\n", string(paymentJSON))

	err = db.Client.LPush(db.RedisCtx, "payment_jobs", paymentJSON).Err()
	if err != nil {
		log.Printf("Error queuing payment: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not queue payment"})
	}

	// retornar status ok
	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "HANDLER", "paymentJSON": paymentJSON, "payment": payment})
}

func PaymentsSummary(ctx *fiber.Ctx) error {
	if ctx.Method() != fiber.MethodGet {
		return ctx.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"error": "Method not allowed",
		})
	}

	fromTime, toTime, filter, err := parseTime(ctx)
	if err != nil {
		if err == http.ErrMissingFile {
			ctx.Status(fiber.StatusBadRequest).SendString("Both 'from' and 'to' query parameters are required, or omit both for all data")
		} else {
			ctx.Status(fiber.StatusBadRequest).SendString("Invalid datetime format")
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(err)
	}

	allPaymentsMap, err := db.Client.HGetAll(db.RedisCtx, "processed_payments").Result()
	if err != nil {
		log.Printf("Erro ao buscar pagamentos processados do Redis: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve processed payments",
		})
	}

	response := model.PaymentsSummaryResponse{
		Default:  model.PaymentsSummary{TotalRequests: 0, TotalAmount: decimal.Zero},
		Fallback: model.PaymentsSummary{TotalRequests: 0, TotalAmount: decimal.Zero},
	}

	for _, paymentJSON := range allPaymentsMap {
		var processedPayment model.ProcessedPayment

		if err := sonic.Unmarshal([]byte(paymentJSON), &processedPayment); err != nil {
			log.Printf("Aviso: falha ao desserializar um registo de pagamento, pulando: %v", err)
			continue
		}
		if filter {
			createdAt, err := time.Parse(time.RFC3339Nano, processedPayment.CreatedAt)
			if err != nil || createdAt.Before(fromTime) || createdAt.After(toTime) {
				continue
			}
		}

		if processedPayment.ProcessedBy == "default" {
			response.Default.TotalRequests++
			response.Default.TotalAmount = response.Default.TotalAmount.Add(processedPayment.Amount)
		} else if processedPayment.ProcessedBy == "fallback" {
			response.Fallback.TotalRequests++
			response.Fallback.TotalAmount = response.Fallback.TotalAmount.Add(processedPayment.Amount)
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func parseTime(ctx *fiber.Ctx) (from, to time.Time, filter bool, err error) {
	fromParam := ctx.Query("from") // Retorna a string do 'from', ou "" se não existir.
	toParam := ctx.Query("to")

	if fromParam == "" && toParam == "" {
		return
	}
	if fromParam == "" || toParam == "" {
		err = http.ErrMissingFile
		return
	}
	from, err = time.Parse(time.RFC3339Nano, fromParam)
	if err != nil {
		return
	}
	to, err = time.Parse(time.RFC3339Nano, toParam)
	if err != nil {
		return
	}
	filter = true
	return
}
