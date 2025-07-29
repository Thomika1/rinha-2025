package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/model"
)

func PaymentProcessor(payment model.Payments, url string) error {

	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment data: %w", err)
	}

	resp, err := http.Post(url+"/payments", "application/json", bytes.NewReader(paymentJSON))
	fmt.Println(paymentJSON)
	requestedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if err != nil {
		return err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("failed to process payment: %s", resp.Status)
		}
	}

	processedPayment := model.ProcessedPayment{
		CorrelationID: payment.CorrelationId,
		Amount:        payment.Amount,
		UrlProcessor:  url,
		CreatedAt:     requestedAt,
	}

	paymentData, err := json.Marshal(processedPayment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment data: %w", err)
	}

	err = db.Client.HSet(db.RedisCtx, "processed_payments", payment.CorrelationId, paymentData).Err()
	if err != nil {
		return fmt.Errorf("failed to save payment in redis: %w", err)
	}

	return nil

}
