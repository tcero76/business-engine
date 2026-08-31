package integration

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tcero76/borrar/postgres/model"
	"gorm.io/gorm"
)

func createTestInventory(t *testing.T, db *gorm.DB, userID uuid.UUID, quantity int, price decimal.Decimal) (uuid.UUID, error) {
	t.Helper()
	product := model.Product{
		SKU:   "TEST-" + uuid.NewString(),
		Name:  "Test Product",
		Price: price,
	}
	if err := db.Create(&product).Error; err != nil {
		return uuid.Nil, fmt.Errorf("create test product: %w", err)
	}
	stock := model.Stock{
		ProductID: product.ID,
		Quantity:  quantity,
	}
	if err := db.Create(&stock).Error; err != nil {
		return uuid.Nil, fmt.Errorf("create test stock: %w", err)
	}
	movement := model.InventoryMovement{
		ProductID:    product.ID,
		UserID:       userID,
		MovementType: model.MovementPurchase,
		Quantity:     quantity,
	}
	if err := db.Create(&movement).Error; err != nil {
		return uuid.Nil, fmt.Errorf("create inventory movement: %w", err)
	}
	return product.ID, nil
}
func cleanupTestInventory(t *testing.T, db *gorm.DB, productID uuid.UUID) error {
	t.Helper()
	if err := db.
		Where("product_id = ?", productID).
		Delete(&model.InventoryMovement{}).Error; err != nil {
		return fmt.Errorf("delete inventory movements: %w", err)
	}
	if err := db.
		Where("product_id = ?", productID).
		Delete(&model.Stock{}).Error; err != nil {
		return fmt.Errorf("delete stock: %w", err)
	}
	if err := db.
		Where("id = ?", productID).
		Delete(&model.Product{}).Error; err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}
func createTestProduct(t *testing.T, db *gorm.DB, price decimal.Decimal) (uuid.UUID, error) {
	t.Helper()
	product := model.Product{
		SKU:   "TEST-" + uuid.NewString(),
		Name:  "Test Product",
		Price: price,
	}

	if err := db.Create(&product).Error; err != nil {
		return uuid.Nil, fmt.Errorf("create test product: %w", err)
	}

	return product.ID, nil
}
func runConcurrent(t *testing.T, n int, fn func() error) []error {
	t.Helper()
	var wg sync.WaitGroup
	errs := make([]error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(index int) {
			defer wg.Done()

			errs[index] = fn()
		}(i)
	}
	wg.Wait()
	return errs
}
func runMixedConcurrent(t *testing.T, n int, first func() error, second func() error) []error {
	t.Helper()
	var wg sync.WaitGroup
	errs := make([]error, n*2)
	wg.Add(n * 2)
	for i := 0; i < n; i++ {
		go func(index int) {
			defer wg.Done()
			errs[index] = first()
		}(i)
		go func(index int) {
			defer wg.Done()

			errs[n+index] = second()
		}(i)
	}
	wg.Wait()
	return errs
}
func createTestProductWithStock(t *testing.T, db *gorm.DB, quantity int, price decimal.Decimal) (uuid.UUID, error) {
	t.Helper()
	product := model.Product{
		SKU:   "TEST-" + uuid.NewString(),
		Name:  "Test Product",
		Price: price,
	}

	if err := db.Create(&product).Error; err != nil {
		return uuid.Nil, fmt.Errorf("create test product: %w", err)
	}

	stock := model.Stock{
		ProductID: product.ID,
		Quantity:  quantity,
	}

	if err := db.Create(&stock).Error; err != nil {
		return uuid.Nil, fmt.Errorf("create test stock: %w", err)
	}

	return product.ID, nil
}
