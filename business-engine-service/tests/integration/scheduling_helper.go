package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tcero76/business-engine/postgres/model"
	"gorm.io/gorm"
)

func createResource(t *testing.T, db *gorm.DB, active bool, price decimal.Decimal) model.Resource {
	t.Helper()
	resource := model.Resource{
		Name:  "Test Resource",
		Price: price,
	}
	require.NoError(t, db.Create(&resource).Error)
	require.NoError(
		t,
		db.Model(&resource).
			Update("active", active).
			Error,
	)
	resource.Active = active
	return resource
}
func createUser(t *testing.T, db *gorm.DB) model.User {
	t.Helper()
	user := model.User{
		ID:       uuid.New(),
		Username: "user-" + uuid.NewString(),
		Email:    uuid.NewString() + "@test.local",
	}
	require.NoError(t, db.Create(&user).Error)
	return user
}
func createAvailability(t *testing.T, db *gorm.DB, resourceID uuid.UUID, start time.Time, end time.Time) model.Availability {
	t.Helper()
	availability := model.Availability{
		ResourceID: resourceID,
		AvailablePeriod: model.TimeRange{
			Lower: start,
			Upper: end,
			Valid: true,
		},
	}
	require.NoError(t, db.Create(&availability).Error)
	return availability
}
func createAppointment(t *testing.T, db *gorm.DB, resourceID uuid.UUID, userID uuid.UUID, start time.Time, end time.Time, status model.AppointmentStatus) model.Appointment {
	t.Helper()
	appointment := model.Appointment{
		ResourceID: resourceID,
		UserID:     userID,
		AppointmentPeriod: model.TimeRange{
			Lower: start,
			Upper: end,
			Valid: true,
		},
		Status: status,
	}
	require.NoError(t, db.Create(&appointment).Error)
	return appointment
}
func deleteAvailability(t *testing.T, db *gorm.DB, resourceID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("resource_id = ?", resourceID).
			Delete(&model.Availability{}).Error,
	)
}
func deleteResource(t *testing.T, db *gorm.DB, resourceID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("id = ?", resourceID).
			Delete(&model.Resource{}).Error,
	)
}
func deleteUser(t *testing.T, db *gorm.DB, userID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("id = ?", userID).
			Delete(&model.User{}).Error,
	)
}
func deleteAppointments(t *testing.T, db *gorm.DB, resourceID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("resource_id = ?", resourceID).
			Delete(&model.Appointment{}).Error,
	)
}
func deleteAppointment(t *testing.T, db *gorm.DB, appointmentID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("id = ?", appointmentID).
			Delete(&model.Appointment{}).Error,
	)
}
