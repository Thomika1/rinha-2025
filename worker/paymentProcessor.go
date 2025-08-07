package worker

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bytedance/sonic"

	"github.com/Thomika1/rinha-2025.git/db"
	"github.com/Thomika1/rinha-2025.git/model"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        500, // Máximo de conexões ociosas
		MaxConnsPerHost:     500, // Máximo de conexões por host
		MaxIdleConnsPerHost: 100, // Máximo de conexões ociosas por host
		DisableKeepAlives:   false,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
	Timeout: 10 * time.Second, // Timeout para a requisição inteira
}

func PaymentProcessor(payment model.Payments) error {

	statusDefault, err := getHealthFromRedis(db.RedisCtx, db.Client, "health:processor:default")
	if err != nil {
		return fmt.Errorf("could not retrieve health state")
	}
	statusFallback, err := getHealthFromRedis(db.RedisCtx, db.Client, "health:processor:fallback")
	if err != nil {
		return fmt.Errorf("could not retrieve health state")
	}

	ProcessorURL := os.Getenv("PROCESSOR_DEFAULT_URL")
	processedBy := "default"
	if statusDefault.Failing || statusDefault.MinResponseTime > statusFallback.MinResponseTime+300 {
		ProcessorURL = os.Getenv("PROCESSOR_FALLBACK_URL")
		processedBy = "fallback"
	}
	if statusFallback.Failing && statusDefault.Failing {
		return fmt.Errorf("both processors failing ")
	}

	requestedAt := time.Now().UTC().Format(time.RFC3339Nano)

	body := map[string]interface{}{
		"correlationId": payment.CorrelationId,
		"amount":        payment.Amount,
		"requestedAt":   requestedAt,
	}
	bodyJSON, _ := sonic.Marshal(body)
	//fmt.Println("PROCESSOR " + string(bodyJSON))

	resp, err := httpClient.Post(ProcessorURL+"/payments", "application/json", bytes.NewReader(bodyJSON))

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
		ProcessedBy:   processedBy,
		CreatedAt:     requestedAt,
	}

	paymentData, err := sonic.Marshal(processedPayment)
	if err != nil {
		return fmt.Errorf("falha ao serializar dados do pagamento processado: %w", err)
	}
	err = db.Client.HSet(db.RedisCtx, "processed_payments", processedPayment.CorrelationID, paymentData).Err()
	if err != nil {
		return fmt.Errorf("falha ao guardar pagamento processado no Redis: %w", err)
	}

	return nil
}
