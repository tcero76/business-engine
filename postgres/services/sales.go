package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/tcero76/borrar/borrar-service/dto"
	"github.com/tcero76/borrar/postgres/model"
)

type SalesService struct {
	db *gorm.DB
}

func NewSalesService(db *gorm.DB) *SalesService {
	return &SalesService{
		db: db,
	}
}

func (s *SalesService) GetOrder(ctx context.Context, req dto.GetOrderRequest) (dto.GetOrderResponse, error) {
	var order model.Order
	err := s.db.WithContext(ctx).
		Preload("Items").
		First(&order, "id = ?", req.OrderID).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.GetOrderResponse{}, ErrOrderNotFound
		}
		return dto.GetOrderResponse{}, err
	}
	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, dto.OrderItemResponse{
			ID:            item.ID,
			AppointmentID: item.AppointmentID,
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TotalAmount:   item.TotalAmount,
		})
	}
	return dto.GetOrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		Currency:    order.Currency,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
		Items:       items,
	}, nil
}

func (s *SalesService) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (dto.CreateOrderResponse, error) {
	if req.UserID == uuid.Nil {
		return dto.CreateOrderResponse{}, ErrInvalidUser
	}
	if err := validateCurrency(req.Currency); err != nil {
		return dto.CreateOrderResponse{}, err
	}
	var user model.User
	err := s.db.WithContext(ctx).
		Where("id = ?", req.UserID).
		First(&user).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CreateOrderResponse{}, ErrUserNotFound
	}
	if err != nil {
		return dto.CreateOrderResponse{}, err
	}
	order := model.Order{
		UserID:      req.UserID,
		Status:      model.OrderPending,
		TotalAmount: decimal.Zero,
		Currency:    req.Currency,
	}
	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return dto.CreateOrderResponse{}, err
	}
	return dto.CreateOrderResponse{
		OrderID: order.ID,
		Status:  string(order.Status),
	}, nil
}

func validateCurrency(currency string) error {
	if len(currency) != 3 {
		return ErrInvalidCurrency
	}

	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return ErrInvalidCurrency
		}
	}

	return nil
}

func (s *SalesService) AddProduct(ctx context.Context, req dto.AddProductRequest) (dto.AddProductResponse, error) {
	if req.Quantity <= 0 {
		return dto.AddProductResponse{}, ErrInvalidQuantity
	}

	var response dto.AddProductResponse

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.Order

		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, "id = ?", req.OrderID).
			Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}

			return err
		}

		if order.Status != model.OrderPending {
			return ErrOrderNotPending
		}

		var product model.Product

		err = tx.
			First(&product, "id = ?", req.ProductID).
			Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProductNotFound
			}

			return err
		}

		if !product.Active {
			return ErrProductDisabled
		}

		total := product.Price.Mul(
			decimal.NewFromInt(int64(req.Quantity)),
		)

		item := model.OrderItem{
			OrderID:     order.ID,
			ProductID:   &product.ID,
			Quantity:    req.Quantity,
			UnitPrice:   product.Price,
			TotalAmount: total,
		}

		if err := tx.Create(&item).Error; err != nil {
			return err
		}

		order.TotalAmount = order.TotalAmount.Add(total)

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		response = dto.AddProductResponse{
			OrderItemID: item.ID,
			UnitPrice:   item.UnitPrice,
			TotalAmount: item.TotalAmount,
		}

		return nil
	})

	if err != nil {
		return dto.AddProductResponse{}, err
	}

	return response, nil
}

func (s *SalesService) AddAppointment(ctx context.Context, req dto.AddAppointmentRequest) (dto.AddAppointmentResponse, error) {
	var response dto.AddAppointmentResponse
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.Order
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, "id = ?", req.OrderID).
			Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if order.Status != model.OrderPending {
			return ErrOrderNotPending
		}
		var appointment model.Appointment
		err = tx.
			Preload("Resource").
			First(&appointment, "id = ?", req.AppointmentID).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAppointmentNotFound
			}
			return err
		}
		if appointment.UserID != order.UserID {
			return ErrAppointmentUserMismatch
		}
		if appointment.Status != model.AppointmentConfirmed {
			return ErrAppointmentNotAvailable
		}
		price := appointment.Resource.Price
		item := model.OrderItem{
			OrderID:       order.ID,
			AppointmentID: &appointment.ID,
			Quantity:      1,
			UnitPrice:     price,
			TotalAmount:   price,
		}
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		order.TotalAmount = order.TotalAmount.Add(price)
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		response = dto.AddAppointmentResponse{
			OrderItemID: item.ID,
			UnitPrice:   item.UnitPrice,
			TotalAmount: item.TotalAmount,
		}
		return nil
	})
	if err != nil {
		return dto.AddAppointmentResponse{}, err
	}
	return response, nil
}

func (s *SalesService) CancelOrder(ctx context.Context, req dto.CancelOrderRequest) (dto.CancelOrderResponse, error) {
	var response dto.CancelOrderResponse
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.Order
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, "id = ?", req.OrderID).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if order.Status != model.OrderPending {
			return ErrOrderNotPending
		}
		order.Status = model.OrderCancelled
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		response = dto.CancelOrderResponse{
			OrderID: order.ID,
			Status:  string(order.Status),
		}
		return nil
	})
	if err != nil {
		return dto.CancelOrderResponse{}, err
	}
	return response, nil
}

func (s *SalesService) CreateReturn(ctx context.Context, req dto.CreateReturnRequest) (dto.CreateReturnResponse, error) {
	var response dto.CreateReturnResponse
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.Order
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, "id = ?", req.OrderID).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if order.Status != model.OrderCompleted {
			return ErrOrderNotCompleted
		}
		if len(req.Items) == 0 {
			return ErrInvalidReturn
		}
		returnOrder := model.ReturnOrder{
			OrderID:     order.ID,
			UserID:      req.UserID,
			Status:      model.ReturnOrderPending,
			TotalAmount: decimal.Zero,
			Currency:    order.Currency,
		}
		if err := tx.Create(&returnOrder).Error; err != nil {
			return err
		}
		for _, reqItem := range req.Items {
			if reqItem.Quantity <= 0 {
				return ErrInvalidQuantity
			}
			var orderItem model.OrderItem
			err := tx.
				First(&orderItem, "id = ? AND order_id = ?", reqItem.OrderItemID, order.ID).
				Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrOrderItemNotFound
				}
				return err
			}
			if reqItem.Quantity > orderItem.Quantity {
				return ErrInvalidReturnQuantity
			}
			unitPrice := orderItem.UnitPrice
			total := unitPrice.Mul(
				decimal.NewFromInt(int64(reqItem.Quantity)),
			)
			returnItem := model.ReturnOrderItem{
				ReturnOrderID: returnOrder.ID,
				OrderItemID:   orderItem.ID,
				Quantity:      reqItem.Quantity,
				UnitPrice:     unitPrice,
				TotalAmount:   total,
			}
			if err := tx.Create(&returnItem).Error; err != nil {
				return err
			}
			returnOrder.TotalAmount = returnOrder.TotalAmount.Add(total)
		}
		if err := tx.Save(&returnOrder).Error; err != nil {
			return err
		}
		response = dto.CreateReturnResponse{
			ReturnOrderID: returnOrder.ID,
			TotalAmount:   returnOrder.TotalAmount,
			Currency:      returnOrder.Currency,
		}
		return nil
	})
	if err != nil {
		return dto.CreateReturnResponse{}, err
	}
	return response, nil
}
