package grpc

import (
	"context"
	"errors"
	"time"

	"order-service/model"
	"order-service/repository"
	"order-service/service"
	orderv1 "proto/order/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCServer struct {
	orderv1.UnimplementedOrderServiceServer
	orderSvc service.OrderService
}

func NewOrderGRPCServer(orderSvc service.OrderService) *OrderGRPCServer {
	return &OrderGRPCServer{
		orderSvc: orderSvc,
	}
}

func toProtoOrder(o *model.Order) *orderv1.Order {
	if o == nil {
		return nil
	}

	var protoItems []*orderv1.OrderItem
	for _, it := range o.Items {
		protoItems = append(protoItems, &orderv1.OrderItem{
			Id:          int64(it.ID),
			OrderId:     int64(it.OrderID),
			ProductId:   int64(it.ProductID),
			ProductName: it.ProductName,
			Quantity:    int32(it.Quantity),
			Price:       it.Price,
		})
	}

	var protoAddr *orderv1.ShippingAddress
	if o.Address != nil {
		protoAddr = &orderv1.ShippingAddress{
			Id:            int64(o.Address.ID),
			OrderId:       int64(o.Address.OrderID),
			RecipientName: o.Address.RecipientName,
			PhoneNumber:   o.Address.PhoneNumber,
			AddressLine1:  o.Address.AddressLine1,
			AddressLine2:  o.Address.AddressLine2,
			City:          o.Address.City,
			State:         o.Address.State,
			PostalCode:    o.Address.PostalCode,
			Country:       o.Address.Country,
			CreatedAt:     o.Address.CreatedAt.Format(time.RFC3339),
		}
	}

	return &orderv1.Order{
		Id:              int64(o.ID),
		UserId:          int64(o.UserID),
		Status:          o.Status,
		TotalAmount:     o.TotalAmount,
		ShippingAddress: o.ShippingAddress,
		Items:           protoItems,
		CreatedAt:       o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       o.UpdatedAt.Format(time.RFC3339),
		Address:         protoAddr,
	}
}

func (s *OrderGRPCServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	if req == nil || req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}
	if len(req.GetItems()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "order must contain at least one item")
	}

	var items []model.OrderItemRequest
	for _, it := range req.GetItems() {
		items = append(items, model.OrderItemRequest{
			ProductID:   int(it.GetProductId()),
			Quantity:    int(it.GetQuantity()),
			Price:       it.GetPrice(),
			ProductName: it.GetProductName(),
		})
	}

	var addr *model.ShippingAddress
	if req.GetAddress() != nil {
		pa := req.GetAddress()
		addr = &model.ShippingAddress{
			RecipientName: pa.GetRecipientName(),
			PhoneNumber:   pa.GetPhoneNumber(),
			AddressLine1:  pa.GetAddressLine1(),
			AddressLine2:  pa.GetAddressLine2(),
			City:          pa.GetCity(),
			State:         pa.GetState(),
			PostalCode:    pa.GetPostalCode(),
			Country:       pa.GetCountry(),
		}
	}

	order, err := s.orderSvc.CreateOrder(ctx, int(req.GetUserId()), model.CreateOrderRequest{
		Items:           items,
		ShippingAddress: req.GetShippingAddress(),
		Address:         addr,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create order: "+err.Error())
	}

	return &orderv1.CreateOrderResponse{
		Order: toProtoOrder(order),
	}, nil
}

func (s *OrderGRPCServer) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid order id is required")
	}

	order, err := s.orderSvc.GetOrderByID(ctx, int(req.GetId()))
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "failed to get order: "+err.Error())
	}

	return &orderv1.GetOrderResponse{
		Order: toProtoOrder(order),
	}, nil
}

func (s *OrderGRPCServer) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	if req == nil || req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}

	orders, err := s.orderSvc.GetOrdersByUserID(ctx, int(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list orders: "+err.Error())
	}

	var protoOrders []*orderv1.Order
	for _, o := range orders {
		orderCopy := o
		protoOrders = append(protoOrders, toProtoOrder(&orderCopy))
	}

	return &orderv1.ListOrdersResponse{
		Orders: protoOrders,
	}, nil
}
