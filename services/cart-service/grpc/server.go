package grpc

import (
	"context"
	"errors"
	"time"

	"cart-service/model"
	"cart-service/service"
	cartv1 "proto/cart/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartGRPCServer struct {
	cartv1.UnimplementedCartServiceServer
	cartSvc service.CartService
}

func NewCartGRPCServer(cartSvc service.CartService) *CartGRPCServer {
	return &CartGRPCServer{
		cartSvc: cartSvc,
	}
}

func toProtoCart(c *model.Cart) *cartv1.Cart {
	if c == nil {
		return &cartv1.Cart{
			Items:      []*cartv1.CartItem{},
			TotalPrice: 0,
		}
	}

	var protoItems []*cartv1.CartItem
	for _, it := range c.Items {
		protoItems = append(protoItems, &cartv1.CartItem{
			ProductId:   int64(it.ProductID),
			Quantity:    int32(it.Quantity),
			ProductName: it.ProductName,
			Price:       it.Price,
		})
	}

	return &cartv1.Cart{
		UserId:     int64(c.UserID),
		Items:      protoItems,
		TotalPrice: c.TotalPrice,
		UpdatedAt:  time.Now().Format(time.RFC3339),
	}
}

func (s *CartGRPCServer) GetCart(ctx context.Context, req *cartv1.GetCartRequest) (*cartv1.GetCartResponse, error) {
	if req == nil || req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}

	cart, err := s.cartSvc.GetUserCart(ctx, int(req.GetUserId()))
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to get cart: "+err.Error())
	}

	return &cartv1.GetCartResponse{
		Cart: toProtoCart(cart),
	}, nil
}

func (s *CartGRPCServer) AddToCart(ctx context.Context, req *cartv1.AddToCartRequest) (*cartv1.AddToCartResponse, error) {
	if req == nil || req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}
	if req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid product_id is required")
	}
	if req.GetQuantity() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}

	cart, err := s.cartSvc.AddToCart(ctx, int(req.GetUserId()), model.AddToCartRequest{
		ProductID:   int(req.GetProductId()),
		Quantity:    int(req.GetQuantity()),
		ProductName: req.GetProductName(),
		Price:       req.GetPrice(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to add item to cart: "+err.Error())
	}

	return &cartv1.AddToCartResponse{
		Cart: toProtoCart(cart),
	}, nil
}

func (s *CartGRPCServer) UpdateCartItem(ctx context.Context, req *cartv1.UpdateCartItemRequest) (*cartv1.UpdateCartItemResponse, error) {
	if req == nil || req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}
	if req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid product_id is required")
	}

	cart, err := s.cartSvc.UpdateItemQuantity(ctx, int(req.GetUserId()), int(req.GetProductId()), model.UpdateItemQuantityRequest{
		Quantity: int(req.GetQuantity()),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update cart item: "+err.Error())
	}

	return &cartv1.UpdateCartItemResponse{
		Cart: toProtoCart(cart),
	}, nil
}

func (s *CartGRPCServer) RemoveFromCart(ctx context.Context, req *cartv1.RemoveFromCartRequest) (*cartv1.RemoveFromCartResponse, error) {
	if req == nil || req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}
	if req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid product_id is required")
	}

	cart, err := s.cartSvc.RemoveFromCart(ctx, int(req.GetUserId()), int(req.GetProductId()))
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to remove item from cart: "+err.Error())
	}

	return &cartv1.RemoveFromCartResponse{
		Cart: toProtoCart(cart),
	}, nil
}

func (s *CartGRPCServer) ClearCart(ctx context.Context, req *cartv1.ClearCartRequest) (*cartv1.ClearCartResponse, error) {
	if req == nil || req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}

	err := s.cartSvc.ClearUserCart(ctx, int(req.GetUserId()))
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to clear cart: "+err.Error())
	}

	return &cartv1.ClearCartResponse{
		Success: true,
	}, nil
}
