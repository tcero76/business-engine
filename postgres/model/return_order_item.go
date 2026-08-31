package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ReturnOrderItem struct {
	ID            uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ReturnOrderID uuid.UUID       `gorm:"type:uuid;not null"`
	OrderItemID   uuid.UUID       `gorm:"type:uuid;not null"`
	Quantity      int             `gorm:"not null"`
	UnitPrice     decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	TotalAmount   decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	CreatedAt     time.Time       `gorm:"not null;default:now()"`

	ReturnOrder ReturnOrder `gorm:"foreignKey:ReturnOrderID;references:ID"`
	OrderItem   OrderItem   `gorm:"foreignKey:OrderItemID;references:ID"`
}

func (ReturnOrderItem) TableName() string {
	return "sales.return_order_items"
}