package model

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/shopspring/decimal"
)

var FallbackURL string = os.Getenv("PROCESSOR_FALLBACK_URL")
var DefaultURL string = os.Getenv("PROCESSOR_DEFAULT_URL")

var HttpClient = &http.Client{
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

type Payments struct {
	CorrelationId string          `json:"correlationId"`
	Amount        decimal.Decimal `json:"amount"`
}

type PaymentsSummary struct {
	TotalRequests int64           `json:"totalRequests"`
	TotalAmount   decimal.Decimal `json:"totalAmount"`
}

type PaymentsSummaryResponse struct {
	Default  PaymentsSummary `json:"default"`
	Fallback PaymentsSummary `json:"fallback"`
}

type ProcessedPayment struct {
	CorrelationID string          `json:"correlationId"`
	Amount        decimal.Decimal `json:"amount"`
	ProcessedBy   string          `json:"processor"`
	CreatedAt     string          `json:"requestedAt"`
}
