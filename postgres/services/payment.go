package services

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/shopspring/decimal"
	"github.com/tcero76/borrar/borrar-service/dto"
	"github.com/tcero76/borrar/postgres/model"
)

type PaymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		db: db,
	}
}

func (s *PaymentService) GetPayment(
	ctx context.Context,
	req dto.GetPaymentRequest,
) (dto.GetPaymentResponse, error) {
	var payment model.Payment

	err := s.db.WithContext(ctx).
		First(&payment, "id = ?", req.PaymentID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.GetPaymentResponse{}, ErrPaymentNotFound
		}

		return dto.GetPaymentResponse{}, err
	}

	return dto.GetPaymentResponse{
		Payment: dto.PaymentDTO{
			ID:        payment.ID,
			OrderID:   payment.OrderID,
			Amount:    payment.Amount,
			Currency:  payment.Currency,
			Status:    string(payment.Status),
			CreatedAt: payment.CreatedAt,
			UpdatedAt: payment.UpdatedAt,
		},
	}, nil
}

func (s *PaymentService) Pay(
	ctx context.Context,
	req dto.PayRequest,
) (dto.PayResponse, error) {
	var response dto.PayResponse

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

		if order.TotalAmount.LessThanOrEqual(decimal.Zero) {
			return ErrInvalidOrder
		}

		payment := model.Payment{
			OrderID:  order.ID,
			Amount:   order.TotalAmount,
			Currency: order.Currency,
			Status:   model.PaymentCompleted,
		}

		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		order.Status = model.OrderPaid

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		response = dto.PayResponse{
			Payment: dto.PaymentDTO{
				ID:        payment.ID,
				OrderID:   payment.OrderID,
				Amount:    payment.Amount,
				Currency:  payment.Currency,
				Status:    string(payment.Status),
				CreatedAt: payment.CreatedAt,
				UpdatedAt: payment.UpdatedAt,
			},
		}

		return nil
	})

	if err != nil {
		return dto.PayResponse{}, err
	}

	return response, nil
}

func (s *PaymentService) Refund(
	ctx context.Context,
	req dto.RefundRequest,
) (dto.RefundResponse, error) {
	var response dto.RefundResponse

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var payment model.Payment

		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&payment, "id = ?", req.PaymentID).
			Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}

			return err
		}

		if payment.Status != model.PaymentCompleted {
			return ErrPaymentNotRefundable
		}

		payment.Status = model.PaymentRefunded

		if err := tx.Save(&payment).Error; err != nil {
			return err
		}

		response = dto.RefundResponse{
			Payment: dto.PaymentDTO{
				ID:        payment.ID,
				OrderID:   payment.OrderID,
				Amount:    payment.Amount,
				Currency:  payment.Currency,
				Status:    string(payment.Status),
				CreatedAt: payment.CreatedAt,
				UpdatedAt: payment.UpdatedAt,
			},
		}

		return nil
	})

	if err != nil {
		return dto.RefundResponse{}, err
	}

	return response, nil
}
