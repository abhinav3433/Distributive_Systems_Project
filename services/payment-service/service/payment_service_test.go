package service

import (
	"context"
	"errors"
	"testing"

	"payment-service/model"
	"payment-service/repository"
)

func setupTestPaymentService() PaymentService {
	repo := repository.NewMockPaymentRepository()
	return NewPaymentService(repo)
}

func TestProcessPaymentSuccess(t *testing.T) {
	svc := setupTestPaymentService()
	ctx := context.Background()

	req := model.ProcessPaymentRequest{
		OrderID:       1001,
		UserID:        50,
		Amount:        299.99,
		PaymentMethod: "VISA_SIMULATED",
	}

	payment, err := svc.ProcessPayment(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error processing payment, got: %v", err)
	}

	if payment.ID == 0 {
		t.Errorf("Expected non-zero payment ID")
	}
	if payment.OrderID != 1001 {
		t.Errorf("Expected OrderID 1001, got %d", payment.OrderID)
	}
	if payment.UserID != 50 {
		t.Errorf("Expected UserID 50, got %d", payment.UserID)
	}
	if payment.Amount != 299.99 {
		t.Errorf("Expected Amount 299.99, got %f", payment.Amount)
	}
	if payment.Status != model.StatusSuccess {
		t.Errorf("Expected Status SUCCESS, got %s", payment.Status)
	}
	if payment.TransactionID == "" {
		t.Errorf("Expected generated TransactionID")
	}
}

func TestProcessPaymentValidation(t *testing.T) {
	svc := setupTestPaymentService()
	ctx := context.Background()

	// 1. Invalid Order ID
	_, err := svc.ProcessPayment(ctx, model.ProcessPaymentRequest{
		OrderID: 0,
		UserID:  1,
		Amount:  100,
	})
	if err == nil {
		t.Errorf("Expected error for zero order ID, got nil")
	}

	// 2. Invalid User ID
	_, err = svc.ProcessPayment(ctx, model.ProcessPaymentRequest{
		OrderID: 1,
		UserID:  0,
		Amount:  100,
	})
	if err == nil {
		t.Errorf("Expected error for zero user ID, got nil")
	}

	// 3. Negative amount
	_, err = svc.ProcessPayment(ctx, model.ProcessPaymentRequest{
		OrderID: 1,
		UserID:  1,
		Amount:  -50,
	})
	if err == nil {
		t.Errorf("Expected error for negative amount, got nil")
	}
}

func TestGetPaymentByOrderID(t *testing.T) {
	svc := setupTestPaymentService()
	ctx := context.Background()

	// Process payment
	_, err := svc.ProcessPayment(ctx, model.ProcessPaymentRequest{
		OrderID: 2002,
		UserID:  75,
		Amount:  450.00,
	})
	if err != nil {
		t.Fatalf("Failed to process payment: %v", err)
	}

	// Fetch payment
	p, err := svc.GetPaymentByOrderID(ctx, 2002)
	if err != nil {
		t.Fatalf("Expected no error fetching payment, got: %v", err)
	}
	if p.OrderID != 2002 {
		t.Errorf("Expected OrderID 2002, got %d", p.OrderID)
	}
	if p.Status != model.StatusSuccess {
		t.Errorf("Expected status SUCCESS, got %s", p.Status)
	}

	// Fetch non-existent payment
	_, err = svc.GetPaymentByOrderID(ctx, 99999)
	if err == nil {
		t.Errorf("Expected error fetching non-existent payment, got nil")
	}
	if !errors.Is(err, repository.ErrPaymentNotFound) {
		t.Errorf("Expected ErrPaymentNotFound, got: %v", err)
	}
}
