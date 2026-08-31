package dto

import "github.com/google/uuid"

type PurchaseRequest struct {
	ProductID uuid.UUID
	UserID    uuid.UUID
	Quantity  int
}

type SellRequest struct {
	ProductID uuid.UUID
	UserID    uuid.UUID
	Quantity  int
}

type AdjustStockRequest struct {
	ProductID      uuid.UUID
	UserID         uuid.UUID
	ActualQuantity int
}

type StockResponse struct {
	ProductID uuid.UUID
	Quantity  int
}

type GetStockRequest struct {
	ProductID uuid.UUID `json:"product_id"`
}