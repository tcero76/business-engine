package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"
	OrderCancelled OrderStatus = "CANCELLED"
	OrderPaid      OrderStatus = "PAID"
	OrderCompleted OrderStatus = "COMPLETED"
)

type Order struct {
	ID          uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      uuid.UUID       `gorm:"type:uuid;not null"`
	Status      OrderStatus     `gorm:"type:varchar(20);not null;default:'PENDING'"`
	TotalAmount decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Currency    string          `gorm:"type:char(3);not null"`
	CreatedAt   time.Time       `gorm:"not null;default:now()"`
	UpdatedAt   time.Time       `gorm:"not null;default:now()"`

	User  User        `gorm:"foreignKey:UserID;references:ID"`
	Items []OrderItem `gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string {
	return "sales.orders"
}