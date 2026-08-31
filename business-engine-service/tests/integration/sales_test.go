package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/tcero76/business-engine/business-engine-service/dto"
	"github.com/tcero76/business-engine/postgres/config"
	"github.com/tcero76/business-engine/postgres/model"
	"github.com/tcero76/business-engine/postgres/services"
)

func TestSalesCreateOrder(t *testing.T) {
	db, err := config.GetPostgres()
	require.NoError(t, err)
	salesService := services.NewSalesService(db)
	t.Run("success", func(t *testing.T) {
		user := createUser(t, db)
		t.Cleanup(func() {
			deleteUser(t, db, user.ID)
		})
		response, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, response.OrderID)

		t.Cleanup(func() {
			deleteOrder(t, db, response.OrderID)
		})

		require.Equal(t, model.OrderPending, model.OrderStatus(response.Status))

		var order model.Order

		require.NoError(
			t,
			db.First(&order, "id = ?", response.OrderID).Error,
		)

		require.Equal(t, user.ID, order.UserID)
		require.Equal(t, model.OrderPending, order.Status)
		require.True(t, order.TotalAmount.Equal(decimal.Zero))
		require.Equal(t, "CLP", order.Currency)
	})
	t.Run("invalid user", func(t *testing.T) {
		response, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   uuid.New(),
				Currency: "CLP",
			},
		)

		require.ErrorIs(t, err, services.ErrUserNotFound)
		require.Equal(t, uuid.Nil, response.OrderID)
	})
	t.Run("invalid currency", func(t *testing.T) {
		user := createUser(t, db)
		t.Cleanup(func() {
			deleteUser(t, db, user.ID)
		})
		response, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "INVALID",
			},
		)
		require.ErrorIs(t, err, services.ErrInvalidCurrency)
		require.Equal(t, uuid.Nil, response.OrderID)
	})
	t.Run("initial status PENDING", func(t *testing.T) {
		user := createUser(t, db)

		t.Cleanup(func() {
			deleteUser(t, db, user.ID)
		})

		response, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)

		require.NoError(t, err)

		t.Cleanup(func() {
			deleteOrder(t, db, response.OrderID)
		})

		var order model.Order

		require.NoError(
			t,
			db.First(&order, "id = ?", response.OrderID).Error,
		)

		require.Equal(t, model.OrderPending, order.Status)
	})
	t.Run("initial total zero", func(t *testing.T) {
		user := createUser(t, db)
		defer deleteUser(t, db, user.ID)

		response, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)

		require.NoError(t, err)

		var order model.Order
		require.NoError(
			t,
			db.First(&order, "id = ?", response.OrderID).Error,
		)

		require.True(t, order.TotalAmount.IsZero())

		require.NoError(
			t,
			db.Delete(&order).Error,
		)
	})
}
func TestSalesGetOrder(t *testing.T) {
	db, err := config.GetPostgres()
	require.NoError(t, err)
	salesService := services.NewSalesService(db)
	t.Run("success", func(t *testing.T) {
		user := createUser(t, db)
		defer deleteUser(t, db, user.ID)

		createResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		defer deleteOrder(t, db, createResponse.OrderID)

		response, err := salesService.GetOrder(
			context.Background(),
			dto.GetOrderRequest{
				OrderID: createResponse.OrderID,
			},
		)

		require.NoError(t, err)

		require.Equal(t, createResponse.OrderID, response.ID)
		require.Equal(t, user.ID, response.UserID)
		require.Equal(t, "PENDING", response.Status)
		require.True(t, response.TotalAmount.IsZero())
		require.Equal(t, "CLP", response.Currency)
		require.Empty(t, response.Items)
	})

	t.Run("not found", func(t *testing.T) {
		response, err := salesService.GetOrder(
			context.Background(),
			dto.GetOrderRequest{
				OrderID: uuid.New(),
			},
		)

		require.ErrorIs(t, err, services.ErrOrderNotFound)
		require.Equal(t, dto.GetOrderResponse{}, response)
	})
}
func TestSalesAddProduct(t *testing.T) {
	db, err := config.GetPostgres()
	require.NoError(t, err)
	salesService := services.NewSalesService(db)
	t.Run("success", func(t *testing.T) {
		user := createUser(t, db)
		productID, err := createTestProduct(t, db, decimal.NewFromInt(100))
		require.NoError(t, err)

		t.Cleanup(func() {
			deleteOrderItemsByProduct(t, db, productID)
			deleteOrdersByUser(t, db, user.ID)
			cleanupTestInventory(t, db, productID)
			deleteUser(t, db, user.ID)
		})

		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		response, err := salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: productID,
				Quantity:  2,
			},
		)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, response.OrderItemID)
		require.True(t, response.UnitPrice.Equal(decimal.NewFromInt(100)))
		require.True(t, response.TotalAmount.Equal(decimal.NewFromInt(200)))

		var item model.OrderItem

		require.NoError(
			t,
			db.First(&item, "id = ?", response.OrderItemID).Error,
		)

		require.Equal(t, orderResponse.OrderID, item.OrderID)
		require.Equal(t, productID, *item.ProductID)
		require.Equal(t, 2, item.Quantity)
		require.True(t, item.UnitPrice.Equal(decimal.NewFromInt(100)))
		require.True(t, item.TotalAmount.Equal(decimal.NewFromInt(200)))
	})
	t.Run("doesn't modify stock", func(t *testing.T) {
		user := createUser(t, db)
		productID, err := createTestProductWithStockAndPrice(
			t,
			db,
			10,
			decimal.NewFromInt(100),
		)
		require.NoError(t, err)

		t.Cleanup(func() {
			deleteOrderItemsByProduct(t, db, productID)
			deleteOrdersByUser(t, db, user.ID)
			cleanupTestInventory(t, db, productID)
			deleteUser(t, db, user.ID)
		})

		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		_, err = salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: productID,
				Quantity:  3,
			},
		)
		require.NoError(t, err)

		var stock model.Stock

		require.NoError(
			t,
			db.First(&stock, "product_id = ?", productID).Error,
		)

		require.Equal(t, 10, stock.Quantity)
	})
	t.Run("uses product price", func(t *testing.T) {
		user := createUser(t, db)
		productID, err := createTestProduct(t, db, decimal.NewFromInt(1500))
		require.NoError(t, err)

		t.Cleanup(func() {
			deleteOrderItemsByProduct(t, db, productID)
			deleteOrdersByUser(t, db, user.ID)
			deleteProduct(t, db, productID)
			deleteUser(t, db, user.ID)
		})

		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		response, err := salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: productID,
				Quantity:  2,
			},
		)
		require.NoError(t, err)

		require.True(
			t,
			response.UnitPrice.Equal(decimal.NewFromInt(1500)),
		)

		require.True(
			t,
			response.TotalAmount.Equal(decimal.NewFromInt(3000)),
		)
	})
	t.Run("updates order total", func(t *testing.T) {
		user := createUser(t, db)
		productID, err := createTestProduct(t, db, decimal.NewFromInt(100))
		require.NoError(t, err)

		t.Cleanup(func() {
			deleteOrderItemsByProduct(t, db, productID)
			deleteOrdersByUser(t, db, user.ID)
			deleteProduct(t, db, productID)
			deleteUser(t, db, user.ID)
		})

		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		_, err = salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: productID,
				Quantity:  2,
			},
		)
		require.NoError(t, err)

		_, err = salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: productID,
				Quantity:  3,
			},
		)
		require.NoError(t, err)

		var order model.Order

		require.NoError(
			t,
			db.First(&order, "id = ?", orderResponse.OrderID).Error,
		)

		require.True(
			t,
			order.TotalAmount.Equal(decimal.NewFromInt(500)),
		)
	})
	t.Run("product not found", func(t *testing.T) {
		user := createUser(t, db)

		t.Cleanup(func() {
			deleteOrdersByUser(t, db, user.ID)
			deleteUser(t, db, user.ID)
		})

		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		_, err = salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: uuid.New(),
				Quantity:  1,
			},
		)

		require.ErrorIs(t, err, services.ErrProductNotFound)
	})
	t.Run("product disabled", func(t *testing.T) {
		user := createUser(t, db)
		productID, err := createTestProduct(t, db, decimal.NewFromInt(100))
		require.NoError(t, err)

		t.Cleanup(func() {
			deleteOrderItemsByProduct(t, db, productID)
			deleteOrdersByUser(t, db, user.ID)
			deleteProduct(t, db, productID)
			deleteUser(t, db, user.ID)
		})

		require.NoError(
			t,
			db.Model(&model.Product{}).
				Where("id = ?", productID).
				Update("active", false).
				Error,
		)

		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		_, err = salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: productID,
				Quantity:  1,
			},
		)

		require.ErrorIs(t, err, services.ErrProductDisabled)
	})
	t.Run("invalid quantity", func(t *testing.T) {
		user := createUser(t, db)
		productID, err := createTestProduct(t, db, decimal.NewFromInt(100))
		require.NoError(t, err)

		t.Cleanup(func() {
			deleteOrderItemsByProduct(t, db, productID)
			deleteOrdersByUser(t, db, user.ID)
			deleteProduct(t, db, productID)
			deleteUser(t, db, user.ID)
		})

		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		_, err = salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: productID,
				Quantity:  0,
			},
		)

		require.ErrorIs(t, err, services.ErrInvalidQuantity)
	})
	t.Run("order not pending", func(t *testing.T) {
		user := createUser(t, db)
		productID, err := createTestProduct(t, db, decimal.NewFromInt(100))
		require.NoError(t, err)

		t.Cleanup(func() {
			deleteOrderItemsByProduct(t, db, productID)
			deleteOrdersByUser(t, db, user.ID)
			deleteProduct(t, db, productID)
			deleteUser(t, db, user.ID)
		})

		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)

		require.NoError(
			t,
			db.Model(&model.Order{}).
				Where("id = ?", orderResponse.OrderID).
				Update("status", model.OrderPaid).
				Error,
		)

		_, err = salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderResponse.OrderID,
				ProductID: productID,
				Quantity:  1,
			},
		)

		require.ErrorIs(t, err, services.ErrOrderNotPending)
	})
}
func TestSalesMixedOrder(t *testing.T) {
	db, err := config.GetPostgres()
	require.NoError(t, err)
	salesService := services.NewSalesService(db)
	t.Run("mixed order", func(t *testing.T) {
		user := createUser(t, db)
		productID, err := createTestProduct(
			t,
			db,
			decimal.NewFromInt(1000),
		)
		require.NoError(t, err)
		resource := createResource(t, db, true, decimal.NewFromInt(1000))
		start := time.Now().Add(24 * time.Hour)
		end := start.Add(time.Hour)
		appointment := createAppointment(
			t,
			db,
			resource.ID,
			user.ID,
			start,
			end,
			model.AppointmentConfirmed,
		)
		var orderID uuid.UUID
		t.Cleanup(func() {
			if orderID != uuid.Nil {
				deleteOrder(t, db, orderID)
			}
			deleteAppointment(t, db, appointment.ID)
			deleteResource(t, db, resource.ID)
			deleteProduct(t, db, productID)
			deleteUser(t, db, user.ID)
		})
		orderResponse, err := salesService.CreateOrder(
			context.Background(),
			dto.CreateOrderRequest{
				UserID:   user.ID,
				Currency: "CLP",
			},
		)
		require.NoError(t, err)
		orderID = orderResponse.OrderID
		_, err = salesService.AddProduct(
			context.Background(),
			dto.AddProductRequest{
				OrderID:   orderID,
				ProductID: productID,
				Quantity:  2,
			},
		)
		require.NoError(t, err)
		_, err = salesService.AddAppointment(
			context.Background(),
			dto.AddAppointmentRequest{
				OrderID:       orderID,
				AppointmentID: appointment.ID,
			},
		)
		require.NoError(t, err)
		var order model.Order
		require.NoError(
			t,
			db.First(&order, "id = ?", orderID).Error,
		)
		require.True(
			t,
			order.TotalAmount.Equal(decimal.NewFromInt(3000)),
		)
		var items []model.OrderItem
		require.NoError(
			t,
			db.Where("order_id = ?", orderID).
				Order("created_at ASC").
				Find(&items).Error,
		)
		require.Len(t, items, 2)

		// Product item
		require.Equal(t, productID, *items[0].ProductID)
		require.Nil(t, items[0].AppointmentID)
		require.Equal(t, 2, items[0].Quantity)
		require.True(
			t,
			items[0].UnitPrice.Equal(decimal.NewFromInt(1000)),
		)
		require.True(
			t,
			items[0].TotalAmount.Equal(decimal.NewFromInt(2000)),
		)

		// Appointment item
		require.Equal(t, appointment.ID, *items[1].AppointmentID)
		require.Nil(t, items[1].ProductID)
		require.Equal(t, 1, items[1].Quantity)
		require.True(
			t,
			items[1].UnitPrice.Equal(decimal.NewFromInt(1000)),
		)
		require.True(
			t,
			items[1].TotalAmount.Equal(decimal.NewFromInt(1000)),
		)
	})
}
func TestSalesCancelOrder(t *testing.T) {
	db, err := config.GetPostgres()
	require.NoError(t, err)
	salesService := services.NewSalesService(db)
	t.Run("success", func(t *testing.T) {
		user := createUser(t, db)

		order := model.Order{
			UserID:      user.ID,
			Status:      model.OrderPending,
			TotalAmount: decimal.NewFromInt(1000),
			Currency:    "CLP",
		}

		require.NoError(t, db.Create(&order).Error)

		t.Cleanup(func() {
			deleteOrder(t, db, order.ID)
			deleteUser(t, db, user.ID)
		})

		response, err := salesService.CancelOrder(
			context.Background(),
			dto.CancelOrderRequest{
				OrderID: order.ID,
			},
		)

		require.NoError(t, err)
		require.Equal(t, order.ID, response.OrderID)
		require.Equal(t, string(model.OrderCancelled), response.Status)

		var cancelledOrder model.Order

		require.NoError(
			t,
			db.First(&cancelledOrder, "id = ?", order.ID).Error,
		)

		require.Equal(t, model.OrderCancelled, cancelledOrder.Status)
	})
	t.Run("not found", func(t *testing.T) {
		response, err := salesService.CancelOrder(
			context.Background(),
			dto.CancelOrderRequest{
				OrderID: uuid.New(),
			},
		)

		require.Error(t, err)
		require.ErrorIs(t, err, services.ErrOrderNotFound)
		require.Equal(t, dto.CancelOrderResponse{}, response)
	})
	t.Run("invalid status", func(t *testing.T) {
		user := createUser(t, db)

		order := model.Order{
			UserID:      user.ID,
			Status:      model.OrderPaid,
			TotalAmount: decimal.NewFromInt(1000),
			Currency:    "CLP",
		}

		require.NoError(t, db.Create(&order).Error)

		t.Cleanup(func() {
			deleteOrder(t, db, order.ID)
			deleteUser(t, db, user.ID)
		})

		response, err := salesService.CancelOrder(
			context.Background(),
			dto.CancelOrderRequest{
				OrderID: order.ID,
			},
		)

		require.Error(t, err)
		require.ErrorIs(t, err, services.ErrOrderNotPending)
		require.Equal(t, dto.CancelOrderResponse{}, response)

		var unchangedOrder model.Order

		require.NoError(
			t,
			db.First(&unchangedOrder, "id = ?", order.ID).Error,
		)

		require.Equal(t, model.OrderPaid, unchangedOrder.Status)
	})
}
func TestSales(t *testing.T) {
	ctx := context.Background()
	db, err := config.GetPostgres()
	require.NoError(t, err)
	salesService := services.NewSalesService(db)
	schedulingService := services.NewSchedulingService(db)
	t.Run("Test general", func(t *testing.T) {
		user := createUser(t, db)
		resource := createResource(t, db, true, decimal.NewFromInt(1000))
		require.Equal(t, true, resource.Active)
		order, err := salesService.CreateOrder(ctx, dto.CreateOrderRequest{
			UserID:   user.ID,
			Currency: "CLP",
		})
		require.NoError(t, err)
		start := time.Now().Add(time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		createAvailability(t, db, resource.ID, start, end)
		scheduleResponse, err := schedulingService.Schedule(ctx, dto.ScheduleRequest{
			CustomerID: user.ID,
			ResourceID: resource.ID,
			StartAt:    start,
			EndAt:      end,
		})
		require.NoError(t, err)
		addAppointmentResponse, err := salesService.AddAppointment(ctx, dto.AddAppointmentRequest{
			OrderID:       order.OrderID,
			AppointmentID: scheduleResponse.Appointment.ID,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			deleteOrder(t, db, order.OrderID)
			deleteAppointment(t, db, scheduleResponse.Appointment.ID)
			deleteAvailability(t, db, resource.ID)
			deleteResource(t, db, resource.ID)
			deleteUser(t, db, user.ID)
		})
		require.True(
			t,
			decimal.NewFromInt(1000).Equal(addAppointmentResponse.TotalAmount),
		)
		require.True(
			t,
			decimal.NewFromInt(1000).Equal(addAppointmentResponse.UnitPrice),
		)

	})
}
