package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username  string    `gorm:"type:varchar(100);not null;uniqueIndex"`
	Email     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (User) TableName() string {
	return "demo.users"
}