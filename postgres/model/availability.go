package model

import (
	"time"

	"github.com/google/uuid"
)

type Availability struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ResourceID      uuid.UUID `gorm:"type:uuid;not null"`
	AvailablePeriod TimeRange `gorm:"type:tstzrange;not null"`
	CreatedAt       time.Time `gorm:"not null;default:now()"`

	Resource Resource `gorm:"foreignKey:ResourceID;references:ID"`
}

func (Availability) TableName() string {
	return "scheduling.availability"
}