package model

import (
	"time"

	"github.com/google/uuid"
)

type MovementType string

const (
	MovementPurchase   MovementType = "PURCHASE"
	MovementSale       MovementType = "SALE"
	MovementAdjustment MovementType = "ADJUSTMENT"
)

type InventoryMovement struct {
	ID           uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ProductID    uuid.UUID    `gorm:"type:uuid;not null"`
	UserID       uuid.UUID    `gorm:"type:uuid;not null"`
	MovementType MovementType `gorm:"type:varchar(20);not null"`
	Quantity     int          `gorm:"not null"`
	CreatedAt    time.Time    `gorm:"not null;default:now()"`

	Product Product `gorm:"foreignKey:ProductID;references:ID"`
	User    User    `gorm:"foreignKey:UserID;references:ID"`
}

func (InventoryMovement) TableName() string {
	return "inventory.inventory_movements"
}