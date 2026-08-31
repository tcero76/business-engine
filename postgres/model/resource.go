package model

import (
	"time"
	"github.com/shopspring/decimal"	
	"github.com/google/uuid"
)

type Resource struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Price     decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Active    bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (Resource) TableName() string {
	return "scheduling.resources"
}