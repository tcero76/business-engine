package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/tcero76/borrar/borrar-service/dto"

	"github.com/tcero76/borrar/postgres/config"
	"github.com/tcero76/borrar/postgres/model"
	"github.com/tcero76/borrar/postgres/services"
)

func TestGetSchedule(t *testing.T) {
	db, err := config.GetPostgres()
	if err != nil {
		t.Logf("Error de conexion: %s", err.Error())
	}
	t.Run("Test scheduling GetSchedule", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user := createUser(t, db)
		start := time.Now().Add(1 * time.Hour).Truncate(time.Second)
		end := start.Add(1 * time.Hour)
		createAppointment(
			t,
			db,
			resource.ID,
			user.ID,
			start,
			end,
			model.AppointmentConfirmed,
		)
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		response, err := service.GetSchedule(
			context.Background(),
			dto.GetScheduleRequest{
				From: start.Add(-30 * time.Minute),
				To:   end.Add(30 * time.Minute),
			},
		)
		require.NoError(t, err)
		require.Len(t, response.Appointments, 1)
		appointment := response.Appointments[0]
		require.Equal(t, resource.ID, appointment.ResourceID)
		require.Equal(t, user.ID, appointment.CustomerID)
		require.Equal(t, start, appointment.StartAt)
		require.Equal(t, end, appointment.EndAt)
		require.Equal(
			t,
			string(model.AppointmentConfirmed),
			appointment.Status,
		)
	})
	t.Run("Test scheduling GetSchedule invalid period", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		now := time.Now()
		_, err = service.GetSchedule(
			context.Background(),
			dto.GetScheduleRequest{
				From: now,
				To:   now.Add(-time.Hour),
			},
		)
		require.ErrorIs(t, err, services.ErrInvalidPeriod)
	})
}
func TestGetAppointment(t *testing.T) {
	db, err := config.GetPostgres()
	if err != nil {
		t.Logf("Error de conexion: %s", err.Error())
	}
	t.Run("Test scheduling GetAppointment", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user := createUser(t, db)
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		created := createAppointment(
			t,
			db,
			resource.ID,
			user.ID,
			start,
			end,
			model.AppointmentConfirmed,
		)
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		response, err := service.GetAppointment(
			context.Background(),
			dto.GetAppointmentRequest{
				AppointmentID: created.ID,
			},
		)

		require.NoError(t, err)

		require.Equal(t, created.ID, response.Appointment.ID)
		require.Equal(t, user.ID, response.Appointment.CustomerID)
		require.Equal(t, resource.ID, response.Appointment.ResourceID)
		require.Equal(t, start, response.Appointment.StartAt)
		require.Equal(t, end, response.Appointment.EndAt)
	})
	t.Run("Test scheduling GetAppointment not found", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		_, err = service.GetAppointment(
			context.Background(),
			dto.GetAppointmentRequest{
				AppointmentID: uuid.New(),
			},
		)

		require.ErrorIs(t, err, services.ErrAppointmentNotFound)
	})
}
func TestSchedule(t *testing.T) {
	ctx := context.Background()
	db, err := config.GetPostgres()
	if err != nil {
		t.Logf("Error de conexion: %s", err.Error())
	}
	t.Run("Test scheduling schedule", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user := createUser(t, db)
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		createAvailability(
			t,
			db,
			resource.ID,
			start.Add(-time.Hour),
			end.Add(time.Hour),
		)
		t.Cleanup(func() {
			deleteAvailability(t, db, resource.ID)
			deleteAppointments(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		response, err := service.Schedule(
			ctx,
			dto.ScheduleRequest{
				CustomerID: user.ID,
				ResourceID: resource.ID,
				StartAt:    start,
				EndAt:      end,
			},
		)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, response.Appointment.ID)
		require.Equal(t, user.ID, response.Appointment.CustomerID)
		require.Equal(t, resource.ID, response.Appointment.ResourceID)
		require.Equal(t, start, response.Appointment.StartAt)
		require.Equal(t, end, response.Appointment.EndAt)
		require.Equal(
			t,
			string(model.AppointmentConfirmed),
			response.Appointment.Status,
		)
	})
	t.Run("Test scheduling schedule invalid period", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		now := time.Now()
		_, err = service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: uuid.New(),
				ResourceID: uuid.New(),
				StartAt:    now,
				EndAt:      now.Add(-time.Hour),
			},
		)
		require.ErrorIs(t, err, services.ErrInvalidPeriod)
	})
	t.Run("Test scheduling schedule resource not found", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		start := time.Now().Add(time.Hour)
		end := start.Add(time.Hour)
		_, err = service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: uuid.New(),
				ResourceID: uuid.New(),
				StartAt:    start,
				EndAt:      end,
			},
		)
		require.ErrorIs(t, err, services.ErrResourceNotFound)
	})
	t.Run("Test scheduling schedule inactive resource", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, false, decimal.NewFromInt(5000))
		user := createUser(t, db)
		// Clean up data created by this test.
		t.Cleanup(func() {
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		start := time.Now().Add(time.Hour)
		end := start.Add(time.Hour)
		createAvailability(
			t,
			db,
			resource.ID,
			start.Add(-time.Hour),
			end.Add(time.Hour),
		)
		t.Cleanup(func() {
			deleteAvailability(t, db, resource.ID)
			deleteAppointments(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		_, err = service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user.ID,
				ResourceID: resource.ID,
				StartAt:    start,
				EndAt:      end,
			},
		)
		require.False(t, resource.Active)
		require.ErrorIs(t, err, services.ErrResourceInactive)
	})
	t.Run("Test scheduling schedule not available", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user := createUser(t, db)
		t.Cleanup(func() {
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		// Availability does not contain the requested period.
		createAvailability(
			t,
			db,
			resource.ID,
			start.Add(-2*time.Hour),
			start.Add(30*time.Minute),
		)
		_, err = service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user.ID,
				ResourceID: resource.ID,
				StartAt:    start,
				EndAt:      end,
			},
		)
		require.ErrorIs(t, err, services.ErrNotAvailable)
	})
	t.Run("Test scheduling schedule conflict", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user1 := createUser(t, db)
		user2 := createUser(t, db)
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user1.ID)
			deleteUser(t, db, user2.ID)
		})
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		createAvailability(
			t,
			db,
			resource.ID,
			start.Add(-time.Hour),
			end.Add(time.Hour),
		)
		_, err = service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user1.ID,
				ResourceID: resource.ID,
				StartAt:    start,
				EndAt:      end,
			},
		)
		require.NoError(t, err)
		_, err = service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user2.ID,
				ResourceID: resource.ID,
				StartAt:    start.Add(30 * time.Minute),
				EndAt:      end.Add(30 * time.Minute),
			},
		)
		require.ErrorIs(t, err, services.ErrAppointmentConflict)
	})
	t.Run("Test scheduling schedule concurrent", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(10))
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)

		var userIDs []uuid.UUID
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			for _, userID := range userIDs {
				deleteUser(t, db, userID)
			}
		})
		createAvailability(
			t,
			db,
			resource.ID,
			start.Add(-time.Hour),
			end.Add(time.Hour),
		)
		const attempts = 10
		var (
			wg        sync.WaitGroup
			mu        sync.Mutex
			successes int
		)
		for i := 0; i < attempts; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				user := createUser(t, db)
				mu.Lock()
				userIDs = append(userIDs, user.ID)
				mu.Unlock()
				_, err := service.Schedule(
					context.Background(),
					dto.ScheduleRequest{
						CustomerID: user.ID,
						ResourceID: resource.ID,
						StartAt:    start,
						EndAt:      end,
					},
				)
				if err == nil {
					mu.Lock()
					successes++
					mu.Unlock()
					return
				}
				require.ErrorIs(t, err, services.ErrAppointmentConflict)
			}()
		}
		wg.Wait()
		require.Equal(t, 1, successes)
	})
}
func TestReschedule(t *testing.T) {
	db, err := config.GetPostgres()
	if err != nil {
		t.Logf("Error de conexion: %s", err.Error())
	}
	t.Run("Test scheduling reschedule", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user := createUser(t, db)
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		newStart := start.Add(2 * time.Hour)
		newEnd := end.Add(2 * time.Hour)
		createAvailability(
			t,
			db,
			resource.ID,
			start.Add(-time.Hour),
			newEnd.Add(time.Hour),
		)
		created, err := service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user.ID,
				ResourceID: resource.ID,
				StartAt:    start,
				EndAt:      end,
			},
		)
		require.NoError(t, err)
		response, err := service.Reschedule(
			context.Background(),
			dto.RescheduleRequest{
				AppointmentID: created.Appointment.ID,
				StartAt:       newStart,
			},
		)
		require.NoError(t, err)
		require.Equal(t, newStart, response.Appointment.StartAt)
		require.Equal(t, newEnd, response.Appointment.EndAt)
		require.Equal(
			t,
			string(model.AppointmentConfirmed),
			response.Appointment.Status,
		)
	})
	t.Run("Test scheduling reschedule conflict", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user1 := createUser(t, db)
		user2 := createUser(t, db)
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user1.ID)
			deleteUser(t, db, user2.ID)
		})
		base := time.Now().Add(time.Hour).Truncate(time.Second)
		createAvailability(
			t,
			db,
			resource.ID,
			base.Add(-time.Hour),
			base.Add(4*time.Hour),
		)
		first, err := service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user1.ID,
				ResourceID: resource.ID,
				StartAt:    base,
				EndAt:      base.Add(time.Hour),
			},
		)
		require.NoError(t, err)
		_, err = service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user2.ID,
				ResourceID: resource.ID,
				StartAt:    base.Add(2 * time.Hour),
				EndAt:      base.Add(3 * time.Hour),
			},
		)
		require.NoError(t, err)

		_, err = service.Reschedule(
			context.Background(),
			dto.RescheduleRequest{
				AppointmentID: first.Appointment.ID,
				StartAt:       base.Add(90 * time.Minute),
			},
		)
		require.ErrorIs(t, err, services.ErrAppointmentConflict)
	})
	t.Run("Test scheduling reschedule cancelled appointment", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user := createUser(t, db)
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		createAvailability(
			t,
			db,
			resource.ID,
			start.Add(-time.Hour),
			end.Add(time.Hour),
		)
		created, err := service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user.ID,
				ResourceID: resource.ID,
				StartAt:    start,
				EndAt:      end,
			},
		)
		require.NoError(t, err)
		_, err = service.Cancel(
			context.Background(),
			dto.CancelRequest{
				AppointmentID: created.Appointment.ID,
			},
		)
		require.NoError(t, err)
		_, err = service.Reschedule(
			context.Background(),
			dto.RescheduleRequest{
				AppointmentID: created.Appointment.ID,
				StartAt:       start.Add(2 * time.Hour),
			},
		)
		require.ErrorIs(t, err, services.ErrInvalidStatus)
	})
}
func TestCancel(t *testing.T) {
	db, err := config.GetPostgres()
	if err != nil {
		t.Logf("Error de conexion: %s", err.Error())
	}
	t.Run("Test scheduling cancel", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user := createUser(t, db)
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		createAvailability(
			t,
			db,
			resource.ID,
			start.Add(-time.Hour),
			end.Add(time.Hour),
		)
		created, err := service.Schedule(
			context.Background(),
			dto.ScheduleRequest{
				CustomerID: user.ID,
				ResourceID: resource.ID,
				StartAt:    start,
				EndAt:      end,
			},
		)
		require.NoError(t, err)
		response, err := service.Cancel(
			context.Background(),
			dto.CancelRequest{
				AppointmentID: created.Appointment.ID,
			},
		)
		require.NoError(t, err)

		require.Equal(
			t,
			string(model.AppointmentCancelled),
			response.Appointment.Status,
		)
	})
	t.Run("Test scheduling cancel already cancelled", func(t *testing.T) {
		service := services.NewSchedulingService(db)
		resource := createResource(t, db, true, decimal.NewFromInt(5000))
		user := createUser(t, db)
		t.Cleanup(func() {
			deleteAppointments(t, db, resource.ID)
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		appointment := createAppointment(
			t,
			db,
			resource.ID,
			user.ID,
			start,
			end,
			model.AppointmentCancelled,
		)
		_, err = service.Cancel(
			context.Background(),
			dto.CancelRequest{
				AppointmentID: appointment.ID,
			},
		)

		require.ErrorIs(t, err, services.ErrInvalidStatus)
	})
}
