package grpc

import (
	"context"
	"testing"

	"payment-service/repository"
	"payment-service/service"
	paymentv1 "proto/payment/v1"
)

func setupTestPaymentGRPC() (*PaymentGRPCServer, service.PaymentService) {
	repo := repository.NewMockPaymentRepository()
	svc := service.NewPaymentService(repo)
	return NewPaymentGRPCServer(svc), svc
}

func TestPaymentGRPC_ProcessAndGet(t *testing.T) {
	server, _ := setupTestPaymentGRPC()
	ctx := context.Background()

	// Process Payment
	req := &paymentv1.ProcessPaymentRequest{
		OrderId:       101,
		UserId:        15,
		Amount:        299.99,
		PaymentMethod: "CREDIT_CARD_SIMULATED",
	}

	procResp, err := server.ProcessPayment(ctx, req)
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	p := procResp.GetPayment()
	if p.GetOrderId() != 101 {
		t.Errorf("expected order id 101, got %d", p.GetOrderId())
	}
	if p.GetStatus() != "SUCCESS" {
		t.Errorf("expected status SUCCESS, got %s", p.GetStatus())
	}
	if p.GetTransactionId() == "" {
		t.Error("expected non-empty transaction id")
	}

	// Get Payment By Order ID
	getResp, err := server.GetPaymentByOrderID(ctx, &paymentv1.GetPaymentRequest{OrderId: 101})
	if err != nil {
		t.Fatalf("GetPaymentByOrderID failed: %v", err)
	}
	if getResp.GetPayment().GetTransactionId() != p.GetTransactionId() {
		t.Errorf("expected transaction ID %s, got %s", p.GetTransactionId(), getResp.GetPayment().GetTransactionId())
	}
}
