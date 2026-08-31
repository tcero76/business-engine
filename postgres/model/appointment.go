package model

import (
	"time"

	"github.com/google/uuid"
)

type AppointmentStatus string

const (
	AppointmentConfirmed AppointmentStatus = "CONFIRMED"
	AppointmentCancelled AppointmentStatus = "CANCELLED"
	AppointmentCompleted AppointmentStatus = "COMPLETED"
)

type Appointment struct {
	ID                uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ResourceID        uuid.UUID         `gorm:"type:uuid;not null"`
	UserID            uuid.UUID         `gorm:"type:uuid;not null"`
	AppointmentPeriod TimeRange         `gorm:"type:tstzrange;not null"`
	Status            AppointmentStatus `gorm:"type:varchar(20);not null;default:'CONFIRMED'"`
	CreatedAt         time.Time         `gorm:"not null;default:now()"`
	UpdatedAt         time.Time         `gorm:"not null;default:now()"`

	Resource Resource `gorm:"foreignKey:ResourceID;references:ID"`
	User     User     `gorm:"foreignKey:UserID;references:ID"`
}

func (Appointment) TableName() string {
	return "scheduling.appointments"
}