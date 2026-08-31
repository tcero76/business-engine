package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetPaymentRequest struct {
	PaymentID uuid.UUID `json:"payment_id"`
}

type GetPaymentResponse struct {
	Payment PaymentDTO `json:"payment"`
}

type PayRequest struct {
	OrderID  uuid.UUID `json:"order_id"`
}

type PayResponse struct {
	Payment PaymentDTO `json:"payment"`
}

type RefundRequest struct {
	PaymentID uuid.UUID `json:"payment_id"`
}

type RefundResponse struct {
	Payment PaymentDTO `json:"payment"`
}

type PaymentDTO struct {
	ID        uuid.UUID       `json:"id"`
	OrderID   uuid.UUID       `json:"order_id"`
	Amount    decimal.Decimal `json:"amount"`
	Currency  string          `json:"currency"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}