package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tcero76/borrar/borrar-service/dto"
)

type Inventory interface {
	GetStock(ctx context.Context, req dto.GetStockRequest) (dto.StockResponse, error)
	Purchase(request dto.PurchaseRequest) error
	Sell(request dto.SellRequest) error
	AdjustStock(request dto.AdjustStockRequest) error
	DisableProduct(productID uuid.UUID) error
	EnableProduct(productID uuid.UUID) error
}

type Scheduling interface {
	GetSchedule(ctx context.Context, req dto.GetScheduleRequest) (dto.GetScheduleResponse, error)
	GetAppointment(ctx context.Context, req dto.GetAppointmentRequest) (dto.GetAppointmentResponse, error)
	Schedule(ctx context.Context, req dto.ScheduleRequest) (dto.ScheduleResponse, error)
	Reschedule(ctx context.Context, req dto.RescheduleRequest) (dto.RescheduleResponse, error)
	Cancel(ctx context.Context, req dto.CancelRequest) (dto.CancelResponse, error)
}

type Payment interface {
	GetPayment(ctx context.Context, req dto.GetPaymentRequest) (dto.GetPaymentResponse, error)
	Pay(ctx context.Context, req dto.PayRequest) (dto.PayResponse, error)
	Refund(ctx context.Context, req dto.RefundRequest) (dto.RefundResponse, error)
}

type Sales interface {
	GetOrder(ctx context.Context, req dto.GetOrderRequest) (dto.GetOrderResponse, error)
	CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (dto.CreateOrderResponse, error)
	AddProduct(ctx context.Context, req dto.AddProductRequest) (dto.AddProductResponse, error)
	AddAppointment(ctx context.Context, req dto.AddAppointmentRequest) (dto.AddAppointmentResponse, error)
	PayOrder(ctx context.Context, req dto.PayOrderRequest) (dto.PayOrderResponse, error)
	CancelOrder(ctx context.Context, req dto.CancelOrderRequest) (dto.CancelOrderResponse, error)
	CreateReturn(ctx context.Context, req dto.CreateReturnRequest) (dto.CreateReturnResponse, error)
}
