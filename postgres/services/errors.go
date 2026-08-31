package services

import "errors"

var (
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")

	// Inventory
	ErrProductNotFound   = errors.New("product not found")
	ErrProductDisabled   = errors.New("product is disabled")
	ErrInsufficientStock = errors.New("insufficient stock")

	// Sales
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderNotPending       = errors.New("order is not pending")
	ErrOrderNotCompleted     = errors.New("order is not completed")
	ErrOrderItemNotFound     = errors.New("order item not found")
	ErrInvalidReturn         = errors.New("invalid return")
	ErrInvalidReturnQuantity = errors.New("invalid return quantity")
	ErrInvalidCurrency       = errors.New("invalid currency")
	ErrInvalidUser           = errors.New("invalid user")

	// Scheduling
	ErrAppointmentNotFound     = errors.New("appointment not found")
	ErrAppointmentNotAvailable = errors.New("appointment is not available")
	ErrAppointmentUserMismatch = errors.New("appointment does not belong to order user")
	ErrInvalidPeriod           = errors.New("invalid appointment period")
	ErrResourceNotFound        = errors.New("resource not found")
	ErrResourceInactive        = errors.New("resource is inactive")
	ErrNotAvailable            = errors.New("resource is not available")
	ErrAppointmentConflict     = errors.New("appointment conflicts with another appointment")
	ErrInvalidStatus           = errors.New("invalid appointment status")

	// payment
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentNotRefundable = errors.New("payment is not refundable")
	ErrInvalidOrder         = errors.New("invalid order")

	ErrUserNotFound = errors.New("user not found")
)
