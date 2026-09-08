package model

import "time"

const (
	StatusPending = "PENDING"
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
)

type Payment struct {
	ID            int       `json:"id"`
	OrderID       int       `json:"order_id"`
	UserID        int       `json:"user_id"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method,omitempty"`
	TransactionID string    `json:"transaction_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProcessPaymentRequest struct {
	OrderID       int     `json:"order_id" binding:"required,gt=0"`
	UserID        int     `json:"user_id"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method"`
}
