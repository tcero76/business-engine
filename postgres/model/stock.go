package model

import (
	"time"

	"github.com/google/uuid"
)

type Stock struct {
	ProductID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Quantity  int       `gorm:"not null;default:0"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`

	Product Product `gorm:"foreignKey:ProductID;references:ID"`
}

func (Stock) TableName() string {
	return "inventory.stock"
}