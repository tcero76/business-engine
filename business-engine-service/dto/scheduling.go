package dto

import (
	"time"

	"github.com/google/uuid"
)

type GetScheduleRequest struct {
	From time.Time
	To   time.Time
}

type GetScheduleResponse struct {
	Appointments []AppointmentDTO
}

type GetAppointmentRequest struct {
	AppointmentID uuid.UUID
}

type GetAppointmentResponse struct {
	Appointment AppointmentDTO
}

type ScheduleRequest struct {
	CustomerID uuid.UUID
	ResourceID uuid.UUID
	StartAt    time.Time
	EndAt      time.Time
}

type ScheduleResponse struct {
	Appointment AppointmentDTO
}

type RescheduleRequest struct {
	AppointmentID uuid.UUID
	StartAt       time.Time
	EndAt         time.Time
}

type RescheduleResponse struct {
	Appointment AppointmentDTO
}

type CancelRequest struct {
	AppointmentID uuid.UUID
}

type CancelResponse struct {
	Appointment AppointmentDTO
}

type AppointmentDTO struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	ResourceID uuid.UUID
	StartAt    time.Time
	EndAt      time.Time
	Status     string
}
