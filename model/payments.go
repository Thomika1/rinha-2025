package model

import (
	"github.com/shopspring/decimal"
)

type Payments struct {
	CorrelationId string          `json:"correlationId"`
	Amount        decimal.Decimal `json:"amount"`
}

type PaymentsSummary struct {
	TotalRequests int             `json:"totalRequests"`
	TotalAmount   decimal.Decimal `json:"totalAmount"`
}

type PaymentsSummaryResponse struct {
	Default  PaymentsSummary `json:"DefaultSummary"`
	Fallback PaymentsSummary `json:"FallbackSummary"`
}

type ServiceHealth struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

type ProcessedPayment struct {
	CorrelationID string          `json:"correlationId"`
	Amount        decimal.Decimal `json:"amount"`
	UrlProcessor  string          `json:"processor"`
	CreatedAt     string          `json:"createdAt"`
}
