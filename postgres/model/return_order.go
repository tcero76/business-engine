package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ReturnOrderStatus string

const (
	ReturnOrderPending   ReturnOrderStatus = "PENDING"
	ReturnOrderCancelled ReturnOrderStatus = "CANCELLED"
	ReturnOrderCompleted ReturnOrderStatus = "COMPLETED"
)

type ReturnOrder struct {
	ID          uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrderID     uuid.UUID         `gorm:"type:uuid;not null"`
	UserID      uuid.UUID         `gorm:"type:uuid;not null"`
	Status      ReturnOrderStatus `gorm:"type:varchar(20);not null;default:'PENDING'"`
	TotalAmount decimal.Decimal   `gorm:"type:numeric(12,2);not null"`
	Currency    string            `gorm:"type:char(3);not null"`
	CreatedAt   time.Time         `gorm:"not null;default:now()"`
	UpdatedAt   time.Time         `gorm:"not null;default:now()"`

	Order Order             `gorm:"foreignKey:OrderID;references:ID"`
	User  User              `gorm:"foreignKey:UserID;references:ID"`
	Items []ReturnOrderItem `gorm:"foreignKey:ReturnOrderID"`
}

func (ReturnOrder) TableName() string {
	return "sales.return_orders"
}