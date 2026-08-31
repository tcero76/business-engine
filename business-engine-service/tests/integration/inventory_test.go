package integration

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tcero76/business-engine/business-engine-service/dto"
	"github.com/tcero76/business-engine/postgres/config"
	"github.com/tcero76/business-engine/postgres/model"
	"github.com/tcero76/business-engine/postgres/services"
)

func TestGetStock(t *testing.T) {
	ctx := context.Background()
	db, err := config.GetPostgres()
	if err != nil {
		log.Fatal("Error de conexión a la BD", err.Error())
	}
	t.Run("Test get stock", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			log.Fatal("Error creating User")
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			log.Fatal("error Creating inventory")
		}
		inventory := services.NewInventoryService(db)
		response, err := inventory.GetStock(ctx, productID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if response.Quantity != 10 {
			t.Fatalf("expected stock 10, got %d", response.Quantity)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
	})
	t.Run("Test GetStock without stock", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		productID := uuid.New()
		response, err := inventory.GetStock(ctx, productID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.ProductID != productID {
			t.Fatalf("expected product ID %v, got %v", productID, response.ProductID)
		}
		if response.Quantity != 0 {
			t.Fatalf("expected stock 0, got %d", response.Quantity)
		}
	})

}
func TestPurchase(t *testing.T) {
	ctx := context.Background()
	db, err := config.GetPostgres()
	if err != nil {
		log.Fatal("Error de conexión a la BD", err.Error())
	}
	t.Run("Test Purchase", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		err = inventory.Purchase(ctx, dto.PurchaseRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  5,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var stock model.Stock
		if err := db.First(&stock, "product_id = ?", productID).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}

		if stock.Quantity != 15 {
			t.Fatalf("expected stock 15, got %d", stock.Quantity)
		}
	})
	t.Run("Test Purchase invalida quantity", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		err = inventory.Purchase(ctx, dto.PurchaseRequest{
			ProductID: uuid.New(),
			UserID:    uuid.New(),
			Quantity:  0,
		})
		if !errors.Is(err, services.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})
	t.Run("Test purchase negative quantity", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		err = inventory.Purchase(ctx, dto.PurchaseRequest{
			ProductID: uuid.New(),
			UserID:    uuid.New(),
			Quantity:  -5,
		})
		if !errors.Is(err, services.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})
	t.Run("Test purchase product not found", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		productID := uuid.New()
		err = inventory.Purchase(ctx, dto.PurchaseRequest{
			ProductID: productID,
			UserID:    uuid.New(),
			Quantity:  5,
		})
		if !errors.Is(err, services.ErrProductNotFound) {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})
	t.Run("Test purchase product disabled", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		if err := db.
			Model(&model.Product{}).
			Where("id = ?", productID).
			Update("active", false).Error; err != nil {
			t.Fatalf("error disabling product: %v", err)
		}
		inventory := services.NewInventoryService(db)
		err = inventory.Purchase(ctx, dto.PurchaseRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  5,
		})
		if !errors.Is(err, services.ErrProductDisabled) {
			t.Fatalf("expected ErrProductDisabled, got %v", err)
		}
	})
	t.Run("Test purchase creates movement", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		const purchaseQuantity = 5
		err = inventory.Purchase(ctx, dto.PurchaseRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  purchaseQuantity,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var movement model.InventoryMovement
		err = db.
			Where("product_id = ?", productID).
			Where("user_id = ?", userID).
			Where("movement_type = ?", model.MovementPurchase).
			Where("quantity = ?", purchaseQuantity).
			First(&movement).Error
		if err != nil {
			t.Fatalf("expected purchase movement: %v", err)
		}
	})
	t.Run("Test purchase without stock", func(t *testing.T) {

		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestProduct(t, db, decimal.NewFromInt(5000))
		if err != nil {
			t.Fatalf("error creating product: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		const purchaseQuantity = 10
		err = inventory.Purchase(ctx, dto.PurchaseRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  purchaseQuantity,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}
		if stock.Quantity != purchaseQuantity {
			t.Fatalf(
				"expected stock %d, got %d",
				purchaseQuantity,
				stock.Quantity,
			)
		}
	})
	t.Run("Test purchase Rollback", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		invalidUserID := uuid.New()
		err = inventory.Purchase(ctx, dto.PurchaseRequest{
			ProductID: productID,
			UserID:    invalidUserID,
			Quantity:  5,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}
		if stock.Quantity != 10 {
			t.Fatalf(
				"expected stock to remain 10 after rollback, got %d",
				stock.Quantity,
			)
		}
	})
}
func TestSell(t *testing.T) {
	ctx := context.Background()
	db, err := config.GetPostgres()
	if err != nil {
		t.Fatalf("error de conexión a la BD: %v", err)
	}
	t.Run("Test sell", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		err = inventory.Sell(ctx, dto.SellRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  3,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}
		if stock.Quantity != 7 {
			t.Fatalf("expected stock 7, got %d", stock.Quantity)
		}
	})
	t.Run("Test sell insufficient stock", func(t *testing.T) {

		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}

		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}

		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})

		inventory := services.NewInventoryService(db)

		err = inventory.Sell(ctx, dto.SellRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  11,
		})

		if !errors.Is(err, services.ErrInsufficientStock) {
			t.Fatalf("expected ErrInsufficientStock, got %v", err)
		}
	})
	t.Run("Test sell without stock", func(t *testing.T) {

		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}

		productID, err := createTestProduct(t, db, decimal.NewFromInt(5000))
		if err != nil {
			t.Fatalf("error creating product: %v", err)
		}

		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})

		inventory := services.NewInventoryService(db)

		err = inventory.Sell(ctx, dto.SellRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  1,
		})

		if !errors.Is(err, services.ErrInsufficientStock) {
			t.Fatalf("expected ErrInsufficientStock, got %v", err)
		}
	})
	t.Run("Test sell invalid quantity", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		err = inventory.Sell(ctx, dto.SellRequest{
			ProductID: uuid.New(),
			UserID:    uuid.New(),
			Quantity:  0,
		})
		if !errors.Is(err, services.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})
	t.Run("Test sell product not found", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		err = inventory.Sell(ctx, dto.SellRequest{
			ProductID: uuid.New(),
			UserID:    uuid.New(),
			Quantity:  1,
		})
		if !errors.Is(err, services.ErrProductNotFound) {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})
	t.Run("Test sell product disabled", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		if err := db.
			Model(&model.Product{}).
			Where("id = ?", productID).
			Update("active", false).Error; err != nil {
			t.Fatalf("error disabling product: %v", err)
		}
		inventory := services.NewInventoryService(db)
		err = inventory.Sell(ctx, dto.SellRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  1,
		})
		if !errors.Is(err, services.ErrProductDisabled) {
			t.Fatalf("expected ErrProductDisabled, got %v", err)
		}
	})
	t.Run("Test sell creates movement", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		const sellQuantity = 3
		inventory := services.NewInventoryService(db)
		err = inventory.Sell(ctx, dto.SellRequest{
			ProductID: productID,
			UserID:    userID,
			Quantity:  sellQuantity,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var movement model.InventoryMovement
		err = db.
			Where("product_id = ?", productID).
			Where("user_id = ?", userID).
			Where("movement_type = ?", model.MovementSale).
			Where("quantity = ?", sellQuantity).
			First(&movement).Error
		if err != nil {
			t.Fatalf("expected sale movement: %v", err)
		}
	})
}
func TestAdjustStock(t *testing.T) {
	ctx := context.Background()
	db, err := config.GetPostgres()
	if err != nil {
		t.Fatalf("error de conexión a la BD: %v", err)
	}
	t.Run("Test adjust stock increase", func(t *testing.T) {

		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      productID,
			UserID:         userID,
			ActualQuantity: 15,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}
		if stock.Quantity != 15 {
			t.Fatalf("expected stock 15, got %d", stock.Quantity)
		}
	})
	t.Run("Test adjust stock decrease", func(t *testing.T) {

		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}

		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}

		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})

		inventory := services.NewInventoryService(db)

		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      productID,
			UserID:         userID,
			ActualQuantity: 6,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var stock model.Stock

		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}

		if stock.Quantity != 6 {
			t.Fatalf("expected stock 6, got %d", stock.Quantity)
		}
	})
	t.Run("Test adjust stock without change", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		var movementsBefore int64
		if err := db.
			Model(&model.InventoryMovement{}).
			Where("product_id = ?", productID).
			Count(&movementsBefore).Error; err != nil {
			t.Fatalf("error counting movements: %v", err)
		}
		inventory := services.NewInventoryService(db)
		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      productID,
			UserID:         userID,
			ActualQuantity: 10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}
		if stock.Quantity != 10 {
			t.Fatalf("expected stock 10, got %d", stock.Quantity)
		}
		var movementsAfter int64
		if err := db.
			Model(&model.InventoryMovement{}).
			Where("product_id = ?", productID).
			Count(&movementsAfter).Error; err != nil {
			t.Fatalf("error counting movements: %v", err)
		}
		if movementsAfter != movementsBefore {
			t.Fatalf(
				"expected %d movements, got %d",
				movementsBefore,
				movementsAfter,
			)
		}
	})
	t.Run("Test adjust stock without stock", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestProduct(t, db, decimal.NewFromInt(5000))
		if err != nil {
			t.Fatalf("error creating product: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		const actualQuantity = 15
		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      productID,
			UserID:         userID,
			ActualQuantity: actualQuantity,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}
		if stock.Quantity != actualQuantity {
			t.Fatalf(
				"expected stock %d, got %d",
				actualQuantity,
				stock.Quantity,
			)
		}
	})
	t.Run("Test adjust stock without stock an zero quantity", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestProduct(t, db, decimal.NewFromInt(5000))
		if err != nil {
			t.Fatalf("error creating product: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      productID,
			UserID:         userID,
			ActualQuantity: 0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting stock: %v", err)
		}
		if stock.Quantity != 0 {
			t.Fatalf("expected stock 0, got %d", stock.Quantity)
		}
		var movementCount int64
		if err := db.
			Model(&model.InventoryMovement{}).
			Where("product_id = ?", productID).
			Count(&movementCount).Error; err != nil {
			t.Fatalf("error counting movements: %v", err)
		}
		if movementCount != 0 {
			t.Fatalf("expected 0 movements, got %d", movementCount)
		}
	})
	t.Run("Test adjust stock invalid quantity", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      uuid.New(),
			UserID:         uuid.New(),
			ActualQuantity: -1,
		})
		if !errors.Is(err, services.ErrInvalidQuantity) {
			t.Fatalf(
				"expected ErrInvalidQuantity, got %v",
				err,
			)
		}
	})
	t.Run("Test adjust stock product not found", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      uuid.New(),
			UserID:         uuid.New(),
			ActualQuantity: 10,
		})
		if !errors.Is(err, services.ErrProductNotFound) {
			t.Fatalf(
				"expected ErrProductNotFound, got %v",
				err,
			)
		}
	})
	t.Run("Test adjust stock product disabled", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		if err := db.
			Model(&model.Product{}).
			Where("id = ?", productID).
			Update("active", false).Error; err != nil {
			t.Fatalf("error disabling product: %v", err)
		}
		inventory := services.NewInventoryService(db)
		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      productID,
			UserID:         userID,
			ActualQuantity: 15,
		})
		if !errors.Is(err, services.ErrProductDisabled) {
			t.Fatalf(
				"expected ErrProductDisabled, got %v",
				err,
			)
		}
	})
	t.Run("Test adjust stock creates absolute movement", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		const actualQuantity = 6
		inventory := services.NewInventoryService(db)
		err = inventory.AdjustStock(ctx, dto.AdjustStockRequest{
			ProductID:      productID,
			UserID:         userID,
			ActualQuantity: actualQuantity,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var movement model.InventoryMovement
		err = db.
			Where("product_id = ?", productID).
			Where("user_id = ?", userID).
			Where("movement_type = ?", model.MovementAdjustment).
			Where("quantity = ?", 4).
			First(&movement).Error
		if err != nil {
			t.Fatalf(
				"expected adjustment movement with quantity 4: %v",
				err,
			)
		}
	})
}
func TestProduct(t *testing.T) {
	ctx := context.Background()
	db, err := config.GetPostgres()
	if err != nil {
		t.Fatalf("error de conexión a la BD: %v", err)
	}
	t.Run("Test disable product", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		err = inventory.DisableProduct(ctx, productID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var product model.Product
		if err := db.
			Where("id = ?", productID).
			First(&product).Error; err != nil {
			t.Fatalf("error getting product: %v", err)
		}
		if product.Active {
			t.Fatal("expected product to be disabled")
		}
	})
	t.Run("Test disable product not found", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		err = inventory.DisableProduct(ctx, uuid.New())
		if !errors.Is(err, services.ErrProductNotFound) {
			t.Fatalf(
				"expected ErrProductNotFound, got %v",
				err,
			)
		}
	})
	t.Run("Test disable product already disabled", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		if err := inventory.DisableProduct(ctx, productID); err != nil {
			t.Fatalf("unexpected error disabling product: %v", err)
		}
		err = inventory.DisableProduct(ctx, productID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		var product model.Product
		if err := db.
			Where("id = ?", productID).
			First(&product).Error; err != nil {
			t.Fatalf("error getting product: %v", err)
		}
		if product.Active {
			t.Fatal("expected product to remain disabled")
		}
	})
	t.Run("Test enabled product", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		if err := db.
			Model(&model.Product{}).
			Where("id = ?", productID).
			Update("active", false).Error; err != nil {
			t.Fatalf("error disabling product: %v", err)
		}
		inventory := services.NewInventoryService(db)
		err = inventory.EnableProduct(ctx, productID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var product model.Product
		if err := db.
			Where("id = ?", productID).
			First(&product).Error; err != nil {
			t.Fatalf("error getting product: %v", err)
		}
		if !product.Active {
			t.Fatal("expected product to be enabled")
		}
	})
	t.Run("Test enable product not found", func(t *testing.T) {
		inventory := services.NewInventoryService(db)
		err = inventory.EnableProduct(ctx, uuid.New())
		if !errors.Is(err, services.ErrProductNotFound) {
			t.Fatalf(
				"expected ErrProductNotFound, got %v",
				err,
			)
		}
	})
	t.Run("Test enable product already enabled", func(t *testing.T) {

		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}

		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		err = inventory.EnableProduct(ctx, productID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		var product model.Product
		if err := db.
			Where("id = ?", productID).
			First(&product).Error; err != nil {
			t.Fatalf("error getting product: %v", err)
		}
		if !product.Active {
			t.Fatal("expected product to remain enabled")
		}
	})
}
func TestConcurrency(t *testing.T) {
	ctx := context.Background()
	db, err := config.GetPostgres()
	if err != nil {
		t.Fatalf("error de conexión a la BD: %v", err)
	}
	t.Run("Test concurrent insufficient stock", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		errs := runConcurrent(t, 100, func() error {
			return inventory.Sell(ctx, dto.SellRequest{
				ProductID: productID,
				UserID:    userID,
				Quantity:  1,
			})
		})
		var successful int
		var insufficientStock int
		for _, err := range errs {
			switch {
			case err == nil:
				successful++

			case errors.Is(err, services.ErrInsufficientStock):
				insufficientStock++

			default:
				t.Fatalf("unexpected error: %v", err)
			}
		}
		if successful != 10 {
			t.Fatalf(
				"expected 10 successful sales, got %d",
				successful,
			)
		}
		if insufficientStock != 90 {
			t.Fatalf(
				"expected 90 insufficient stock errors, got %d",
				insufficientStock,
			)
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting final stock: %v", err)
		}
		if stock.Quantity != 0 {
			t.Fatalf(
				"expected final stock 0, got %d",
				stock.Quantity,
			)
		}
		var movements int64
		if err := db.
			Model(&model.InventoryMovement{}).
			Where("product_id = ?", productID).
			Where("movement_type = ?", model.MovementSale).
			Count(&movements).Error; err != nil {
			t.Fatalf("error counting movements: %v", err)
		}
		if movements != 10 {
			t.Fatalf(
				"expected 10 sale movements, got %d",
				movements,
			)
		}
	})
	t.Run("Test purchase concurrent", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestProductWithStock(t, db, 0, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}
			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		errs := runConcurrent(t, 100, func() error {
			return inventory.Purchase(ctx, dto.PurchaseRequest{
				ProductID: productID,
				UserID:    userID,
				Quantity:  1,
			})
		})
		var successful int
		for _, err := range errs {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			successful++
		}
		if successful != 100 {
			t.Fatalf(
				"expected 100 successful purchases, got %d",
				successful,
			)
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting final stock: %v", err)
		}
		if stock.Quantity != 100 {
			t.Fatalf(
				"expected final stock 100, got %d",
				stock.Quantity,
			)
		}
		var movements int64
		if err := db.
			Model(&model.InventoryMovement{}).
			Where("product_id = ?", productID).
			Where("movement_type = ?", model.MovementPurchase).
			Count(&movements).Error; err != nil {
			t.Fatalf("error counting movements: %v", err)
		}
		if movements != 100 {
			t.Fatalf(
				"expected 100 purchase movements, got %d",
				movements,
			)
		}
	})
	t.Run("Test purchase sell concurrent", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestProductWithStock(t, db, 50, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating product with stock: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		errs := runMixedConcurrent(
			t,
			50,
			func() error {
				return inventory.Purchase(ctx, dto.PurchaseRequest{
					ProductID: productID,
					UserID:    userID,
					Quantity:  1,
				})
			},
			func() error {
				return inventory.Sell(ctx, dto.SellRequest{
					ProductID: productID,
					UserID:    userID,
					Quantity:  1,
				})
			},
		)
		for _, err := range errs {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting final stock: %v", err)
		}
		if stock.Quantity != 50 {
			t.Errorf(
				"expected final stock 50, got %d",
				stock.Quantity,
			)
		}
		var purchases int64
		if err := db.
			Model(&model.InventoryMovement{}).
			Where("product_id = ?", productID).
			Where("movement_type = ?", model.MovementPurchase).
			Count(&purchases).Error; err != nil {
			t.Fatalf("error counting purchase movements: %v", err)
		}
		if purchases != 50 {
			t.Fatalf(
				"expected 50 purchase movements, got %d",
				purchases,
			)
		}
		var sales int64
		if err := db.
			Model(&model.InventoryMovement{}).
			Where("product_id = ?", productID).
			Where("movement_type = ?", model.MovementSale).
			Count(&sales).Error; err != nil {
			t.Fatalf("error counting sale movements: %v", err)
		}
		t.Logf(
			"purchases=%d sales=%d final_stock=%d",
			purchases,
			sales,
			stock.Quantity,
		)
		if sales != 50 {
			t.Fatalf(
				"expected 50 sale movements, got %d",
				sales,
			)
		}
	})
	t.Run("Test AdjustStock concurrent", func(t *testing.T) {
		userID, err := createTestUser(t, db)
		if err != nil {
			t.Fatalf("error creating user: %v", err)
		}
		productID, err := createTestInventory(t, db, userID, 10, decimal.NewFromInt(1000))
		if err != nil {
			t.Fatalf("error creating inventory: %v", err)
		}
		t.Cleanup(func() {
			if err := cleanupTestInventory(t, db, productID); err != nil {
				t.Errorf("error cleaning inventory: %v", err)
			}

			if err := cleanupTestUser(t, db, userID); err != nil {
				t.Errorf("error cleaning user: %v", err)
			}
		})
		inventory := services.NewInventoryService(db)
		errs := runConcurrent(t, 100, func() error {
			return inventory.AdjustStock(ctx, dto.AdjustStockRequest{
				ProductID:      productID,
				UserID:         userID,
				ActualQuantity: 50,
			})
		})
		for _, err := range errs {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		var stock model.Stock
		if err := db.
			Where("product_id = ?", productID).
			First(&stock).Error; err != nil {
			t.Fatalf("error getting final stock: %v", err)
		}
		if stock.Quantity != 50 {
			t.Fatalf(
				"expected final stock 50, got %d",
				stock.Quantity,
			)
		}
		var movements int64
		if err := db.
			Model(&model.InventoryMovement{}).
			Where("product_id = ?", productID).
			Where("movement_type = ?", model.MovementAdjustment).
			Count(&movements).Error; err != nil {
			t.Fatalf("error counting adjustment movements: %v", err)
		}
		if movements != 1 {
			t.Fatalf(
				"expected 1 adjustment movement, got %d",
				movements,
			)
		}
	})
}
