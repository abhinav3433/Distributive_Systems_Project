package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"payment-service/model"
	"payment-service/repository"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type PaymentService interface {
	ProcessPayment(ctx context.Context, req model.ProcessPaymentRequest) (*model.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID int) (*model.Payment, error)
}

type paymentServiceImpl struct {
	repo repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository) PaymentService {
	return &paymentServiceImpl{repo: repo}
}

func generateTransactionID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("txn_sim_%d_%s", time.Now().Unix(), hex.EncodeToString(bytes))
}

func (s *paymentServiceImpl) ProcessPayment(ctx context.Context, req model.ProcessPaymentRequest) (*model.Payment, error) {
	if req.OrderID <= 0 {
		return nil, fmt.Errorf("%w: invalid order ID", ErrInvalidInput)
	}
	if req.UserID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("%w: payment amount must be greater than zero", ErrInvalidInput)
	}

	method := req.PaymentMethod
	if method == "" {
		method = "CREDIT_CARD_SIMULATED"
	}

	// 1. Check if a payment for this order already succeeded
	existing, err := s.repo.GetByOrderID(ctx, req.OrderID)
	if err == nil && existing != nil && existing.Status == model.StatusSuccess {
		return existing, nil
	}

	// 2. Simulate payment processing
	txnID := generateTransactionID()

	payment := &model.Payment{
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Status:        model.StatusSuccess,
		PaymentMethod: method,
		TransactionID: txnID,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to record payment: %w", err)
	}

	return payment, nil
}

func (s *paymentServiceImpl) GetPaymentByOrderID(ctx context.Context, orderID int) (*model.Payment, error) {
	if orderID <= 0 {
		return nil, fmt.Errorf("%w: invalid order ID", ErrInvalidInput)
	}
	return s.repo.GetByOrderID(ctx, orderID)
}
