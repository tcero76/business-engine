package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/tcero76/business-engine/postgres/model"

	"github.com/tcero76/business-engine/business-engine-service/dto"
	IServices "github.com/tcero76/business-engine/business-engine-service/services"
)

type schedulingService struct {
	db *gorm.DB
}

func NewSchedulingService(db *gorm.DB) IServices.Scheduling {
	return &schedulingService{
		db: db,
	}
}

func (s *schedulingService) GetSchedule(ctx context.Context, req dto.GetScheduleRequest) (dto.GetScheduleResponse, error) {
	if !req.From.Before(req.To) {
		return dto.GetScheduleResponse{}, ErrInvalidPeriod
	}
	var appointments []model.Appointment
	err := s.db.WithContext(ctx).
		Where(
			"appointment_period && tstzrange(?, ?, '[)')",
			req.From,
			req.To,
		).
		Order("lower(appointment_period) ASC").
		Find(&appointments).Error
	if err != nil {
		return dto.GetScheduleResponse{}, fmt.Errorf(
			"get schedule: %w",
			err,
		)
	}
	response := dto.GetScheduleResponse{
		Appointments: make([]dto.AppointmentDTO, 0, len(appointments)),
	}

	for _, appointment := range appointments {
		response.Appointments = append(
			response.Appointments,
			appointmentToDTO(appointment),
		)
	}
	return response, nil
}

func (s *schedulingService) GetAppointment(ctx context.Context, req dto.GetAppointmentRequest) (dto.GetAppointmentResponse, error) {
	var appointment model.Appointment
	err := s.db.WithContext(ctx).
		First(&appointment, "id = ?", req.AppointmentID).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.GetAppointmentResponse{}, ErrAppointmentNotFound
	}
	if err != nil {
		return dto.GetAppointmentResponse{}, fmt.Errorf(
			"get appointment: %w",
			err,
		)
	}
	return dto.GetAppointmentResponse{
		Appointment: appointmentToDTO(appointment),
	}, nil
}
func (s *schedulingService) Schedule(ctx context.Context, req dto.ScheduleRequest) (dto.ScheduleResponse, error) {
	if !req.StartAt.Before(req.EndAt) {
		return dto.ScheduleResponse{}, ErrInvalidPeriod
	}
	var appointment model.Appointment
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the resource so concurrent scheduling operations
		// against the same resource are serialized.
		var resource model.Resource
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&resource, "id = ?", req.ResourceID).
			Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResourceNotFound
		}
		if err != nil {
			return fmt.Errorf("get resource: %w", err)
		}
		if !resource.Active {
			return ErrResourceInactive
		}
		period := model.TimeRange{
			Lower: req.StartAt,
			Upper: req.EndAt,
			Valid: true,
		}
		// The requested period must be completely contained
		// inside one availability period.
		var availability model.Availability
		err = tx.
			Where("resource_id = ?", req.ResourceID).
			Where(
				"available_period @> tstzrange(?, ?, '[)')",
				req.StartAt,
				req.EndAt,
			).
			First(&availability).
			Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotAvailable
		}
		if err != nil {
			return fmt.Errorf("check availability: %w", err)
		}
		// Application-level conflict check.
		//
		// PostgreSQL's exclusion constraint remains the final
		// protection against concurrent inserts.
		var conflict int64
		err = tx.
			Model(&model.Appointment{}).
			Where("resource_id = ?", req.ResourceID).
			Where("status = ?", model.AppointmentConfirmed).
			Where(
				"appointment_period && tstzrange(?, ?, '[)')",
				req.StartAt,
				req.EndAt,
			).
			Count(&conflict).
			Error

		if err != nil {
			return fmt.Errorf("check appointment conflict: %w", err)
		}

		if conflict > 0 {
			return ErrAppointmentConflict
		}

		appointment = model.Appointment{
			ResourceID:        req.ResourceID,
			UserID:            req.CustomerID,
			AppointmentPeriod: period,
			Status:            model.AppointmentConfirmed,
		}

		if err := tx.Create(&appointment).Error; err != nil {
			if isExclusionViolation(err) {
				return ErrAppointmentConflict
			}

			return fmt.Errorf("create appointment: %w", err)
		}

		return nil
	})

	if err != nil {
		return dto.ScheduleResponse{}, err
	}

	return dto.ScheduleResponse{
		Appointment: appointmentToDTO(appointment),
	}, nil
}
func (s *schedulingService) Reschedule(
	ctx context.Context,
	req dto.RescheduleRequest,
) (dto.RescheduleResponse, error) {

	var appointment model.Appointment

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&appointment, "id = ?", req.AppointmentID).
			Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAppointmentNotFound
		}

		if err != nil {
			return fmt.Errorf("get appointment: %w", err)
		}

		if appointment.Status != model.AppointmentConfirmed {
			return ErrInvalidStatus
		}

		// Extract the current period.
		currentStart, currentEnd, err := appointmentPeriod(appointment)
		if err != nil {
			return fmt.Errorf("read appointment period: %w", err)
		}

		duration := currentEnd.Sub(currentStart)

		newEnd := req.StartAt.Add(duration)

		if !req.StartAt.Before(newEnd) {
			return ErrInvalidPeriod
		}

		var resource model.Resource

		err = tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&resource, "id = ?", appointment.ResourceID).
			Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResourceNotFound
		}

		if err != nil {
			return fmt.Errorf("get resource: %w", err)
		}

		if !resource.Active {
			return ErrResourceInactive
		}

		// New period must be available.
		var availability model.Availability

		err = tx.
			Where("resource_id = ?", appointment.ResourceID).
			Where(
				"available_period @> tstzrange(?, ?, '[)')",
				req.StartAt,
				newEnd,
			).
			First(&availability).
			Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotAvailable
		}

		if err != nil {
			return fmt.Errorf("check availability: %w", err)
		}

		// Check conflicts, excluding the appointment being rescheduled.
		var conflict int64

		err = tx.
			Model(&model.Appointment{}).
			Where("resource_id = ?", appointment.ResourceID).
			Where("id <> ?", appointment.ID).
			Where("status = ?", model.AppointmentConfirmed).
			Where(
				"appointment_period && tstzrange(?, ?, '[)')",
				req.StartAt,
				newEnd,
			).
			Count(&conflict).
			Error

		if err != nil {
			return fmt.Errorf("check appointment conflict: %w", err)
		}

		if conflict > 0 {
			return ErrAppointmentConflict
		}

		appointment.AppointmentPeriod = model.TimeRange{
			Lower: req.StartAt,
			Upper: newEnd,
			Valid: true,
		}

		if err := tx.Save(&appointment).Error; err != nil {
			if isExclusionViolation(err) {
				return ErrAppointmentConflict
			}

			return fmt.Errorf("reschedule appointment: %w", err)
		}

		return nil
	})

	if err != nil {
		return dto.RescheduleResponse{}, err
	}

	return dto.RescheduleResponse{
		Appointment: appointmentToDTO(appointment),
	}, nil
}

func (s *schedulingService) Cancel(
	ctx context.Context,
	req dto.CancelRequest,
) (dto.CancelResponse, error) {

	var appointment model.Appointment

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&appointment, "id = ?", req.AppointmentID).
			Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAppointmentNotFound
		}

		if err != nil {
			return fmt.Errorf("get appointment: %w", err)
		}

		if appointment.Status != model.AppointmentConfirmed {
			return ErrInvalidStatus
		}

		appointment.Status = model.AppointmentCancelled

		if err := tx.Save(&appointment).Error; err != nil {
			return fmt.Errorf("cancel appointment: %w", err)
		}

		return nil
	})

	if err != nil {
		return dto.CancelResponse{}, err
	}

	return dto.CancelResponse{
		Appointment: appointmentToDTO(appointment),
	}, nil
}

func appointmentToDTO(
	appointment model.Appointment,
) dto.AppointmentDTO {

	startAt := time.Time{}
	endAt := time.Time{}

	if appointment.AppointmentPeriod.Valid {
		startAt = appointment.AppointmentPeriod.Lower
		endAt = appointment.AppointmentPeriod.Upper
	}

	return dto.AppointmentDTO{
		ID:         appointment.ID,
		CustomerID: appointment.UserID,
		ResourceID: appointment.ResourceID,
		StartAt:    startAt,
		EndAt:      endAt,
		Status:     string(appointment.Status),
	}
}

func appointmentPeriod(
	appointment model.Appointment,
) (time.Time, time.Time, error) {
	if !appointment.AppointmentPeriod.Valid {
		return time.Time{}, time.Time{}, ErrInvalidPeriod
	}
	start := appointment.AppointmentPeriod.Lower
	end := appointment.AppointmentPeriod.Upper
	if !start.Before(end) {
		return time.Time{}, time.Time{}, ErrInvalidPeriod
	}
	return start, end, nil
}

func isExclusionViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23P01"
}
