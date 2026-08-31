package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetOrderRequest struct {
	OrderID uuid.UUID `json:"order_id"`
}

type GetOrderResponse struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	Status      string          `json:"status"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	Currency    string          `json:"currency"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	Items []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ID            uuid.UUID       `json:"id"`
	AppointmentID *uuid.UUID      `json:"appointment_id,omitempty"`
	ProductID     *uuid.UUID      `json:"product_id,omitempty"`
	Quantity      int             `json:"quantity"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
}

type CreateOrderRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Currency string    `json:"currency"`
}

type CreateOrderResponse struct {
	OrderID uuid.UUID `json:"order_id"`
	Status  string    `json:"status"`
}

type AddProductRequest struct {
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type AddAppointmentRequest struct {
	OrderID       uuid.UUID `json:"order_id"`
	AppointmentID uuid.UUID `json:"appointment_id"`
}
type AddProductResponse struct {
	OrderItemID uuid.UUID       `json:"order_item_id"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type AddAppointmentResponse struct {
	OrderItemID uuid.UUID       `json:"order_item_id"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type PayOrderRequest struct {
	OrderID uuid.UUID `json:"order_id"`
}

type PayOrderResponse struct {
	OrderID     uuid.UUID       `json:"order_id"`
	Status      string          `json:"status"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	Currency    string          `json:"currency"`
}

type CancelOrderRequest struct {
	OrderID uuid.UUID `json:"order_id"`
}

type CancelOrderResponse struct {
	OrderID uuid.UUID `json:"order_id"`
	Status  string    `json:"status"`
}

type CreateReturnRequest struct {
	OrderID uuid.UUID                 `json:"order_id"`
	UserID  uuid.UUID                 `json:"user_id"`
	Items   []CreateReturnItemRequest `json:"items"`
}

type CreateReturnItemRequest struct {
	OrderItemID uuid.UUID `json:"order_item_id"`
	Quantity    int       `json:"quantity"`
}

type CreateReturnResponse struct {
	ReturnOrderID uuid.UUID       `json:"return_order_id"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	Currency      string          `json:"currency"`
}
