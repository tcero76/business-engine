package integration

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tcero76/business-engine/postgres/model"
	"gorm.io/gorm"
)

func createTestUser(t *testing.T, db *gorm.DB) (uuid.UUID, error) {
	t.Helper()
	ID, err := uuid.NewUUID()
	if err != nil {
		log.Fatal("Error creating User Id")
	}
	user := model.User{
		ID:        ID,
		Username:  "Leonardo",
		CreatedAt: time.Now(),
		Email:     "leonardo.lastra@gmail.com",
	}

	if err := db.Create(&user).Error; err != nil {
		return uuid.Nil, fmt.Errorf("create test user: %w", err)
	}

	return user.ID, nil
}

func cleanupTestUser(t *testing.T, db *gorm.DB, userID uuid.UUID) error {
	t.Helper()
	if err := db.
		Where("id = ?", userID).
		Delete(&model.User{}).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
