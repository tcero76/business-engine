package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "PENDING"
	PaymentCompleted PaymentStatus = "COMPLETED"
	PaymentFailed    PaymentStatus = "FAILED"
	PaymentCancelled PaymentStatus = "CANCELLED"
	PaymentRefunded  PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID       uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrderID  uuid.UUID       `gorm:"type:uuid;not null"`
	Amount   decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Currency string          `gorm:"type:char(3);not null"`
	Status   PaymentStatus   `gorm:"type:varchar(20);not null;default:'PENDING'"`
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`

	Order Order `gorm:"foreignKey:OrderID;references:ID"`
}

func (Payment) TableName() string {
	return "payment.payments"
}