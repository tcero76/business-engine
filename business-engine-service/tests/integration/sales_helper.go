package integration

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tcero76/business-engine/postgres/model"
	"gorm.io/gorm"
)

func createTestProductWithStockAndPrice(t *testing.T, db *gorm.DB, quantity int, price decimal.Decimal) (uuid.UUID, error) {
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
func deleteOrder(t *testing.T, db *gorm.DB, orderID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("order_id = ?", orderID).
			Delete(&model.OrderItem{}).
			Error,
	)

	require.NoError(
		t,
		db.Where("id = ?", orderID).
			Delete(&model.Order{}).
			Error,
	)
}
func deleteProduct(t *testing.T, db *gorm.DB, productID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("id = ?", productID).
			Delete(&model.Product{}).Error,
	)
}
func deleteOrderItemsByProduct(t *testing.T, db *gorm.DB, productID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("product_id = ?", productID).
			Delete(&model.OrderItem{}).Error,
	)
}
func deleteOrdersByUser(t *testing.T, db *gorm.DB, userID uuid.UUID) {
	t.Helper()
	require.NoError(
		t,
		db.Where("user_id = ?", userID).
			Delete(&model.Order{}).Error,
	)
}
