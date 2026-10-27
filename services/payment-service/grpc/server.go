package grpc

import (
	"context"
	"errors"
	"time"

	"payment-service/model"
	"payment-service/repository"
	"payment-service/service"
	paymentv1 "proto/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentGRPCServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	paymentSvc service.PaymentService
}

func NewPaymentGRPCServer(paymentSvc service.PaymentService) *PaymentGRPCServer {
	return &PaymentGRPCServer{
		paymentSvc: paymentSvc,
	}
}

func toProtoPayment(p *model.Payment) *paymentv1.Payment {
	if p == nil {
		return nil
	}
	return &paymentv1.Payment{
		Id:            int64(p.ID),
		OrderId:       int64(p.OrderID),
		UserId:        int64(p.UserID),
		Amount:        p.Amount,
		Status:        p.Status,
		PaymentMethod: p.PaymentMethod,
		TransactionId: p.TransactionID,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *PaymentGRPCServer) GetPaymentByOrderID(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	if req == nil || req.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid order_id is required")
	}

	p, err := s.paymentSvc.GetPaymentByOrderID(ctx, int(req.GetOrderId()))
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, status.Error(codes.NotFound, "payment record not found")
		}
		return nil, status.Error(codes.Internal, "failed to get payment: "+err.Error())
	}

	return &paymentv1.GetPaymentResponse{
		Payment: toProtoPayment(p),
	}, nil
}

func (s *PaymentGRPCServer) ProcessPayment(ctx context.Context, req *paymentv1.ProcessPaymentRequest) (*paymentv1.ProcessPaymentResponse, error) {
	if req == nil || req.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid order_id is required")
	}
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "payment amount must be greater than zero")
	}

	p, err := s.paymentSvc.ProcessPayment(ctx, model.ProcessPaymentRequest{
		OrderID:       int(req.GetOrderId()),
		UserID:        int(req.GetUserId()),
		Amount:        req.GetAmount(),
		PaymentMethod: req.GetPaymentMethod(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to process payment: "+err.Error())
	}

	return &paymentv1.ProcessPaymentResponse{
		Payment: toProtoPayment(p),
	}, nil
}
