package worker

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"

	"github.com/Thomika1/rinha-2025.git/model"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100, // Máximo de conexões ociosas
		MaxConnsPerHost:     100, // Máximo de conexões por host
		MaxIdleConnsPerHost: 100, // Máximo de conexões ociosas por host
		IdleConnTimeout:     90 * time.Second,
	},
	Timeout: 10 * time.Second, // Timeout para a requisição inteira
}

func PaymentProcessor(payment model.Payments, url string) error {

	paymentJSON, err := sonic.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment data: %w", err)
	}

	resp, err := httpClient.Post(url+"/payments", "application/json", bytes.NewReader(paymentJSON))
	//fmt.Println(paymentJSON)
	//requestedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if err != nil {
		return err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("failed to process payment: %s", resp.Status)
		}
	}

	return nil

}
