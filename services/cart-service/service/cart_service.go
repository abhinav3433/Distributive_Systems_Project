package service

import (
	"context"
	"errors"
	"fmt"

	"cart-service/model"
	"cart-service/repository"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type CartService interface {
	GetUserCart(ctx context.Context, userID int) (*model.Cart, error)
	AddToCart(ctx context.Context, userID int, req model.AddToCartRequest) (*model.Cart, error)
	UpdateItemQuantity(ctx context.Context, userID int, productID int, req model.UpdateItemQuantityRequest) (*model.Cart, error)
	RemoveFromCart(ctx context.Context, userID int, productID int) (*model.Cart, error)
	ClearUserCart(ctx context.Context, userID int) error
}

type cartServiceImpl struct {
	repo repository.CartRepository
}

func NewCartService(repo repository.CartRepository) CartService {
	return &cartServiceImpl{repo: repo}
}

func (s *cartServiceImpl) GetUserCart(ctx context.Context, userID int) (*model.Cart, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	return s.repo.GetCart(ctx, userID)
}

func (s *cartServiceImpl) AddToCart(ctx context.Context, userID int, req model.AddToCartRequest) (*model.Cart, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	if req.ProductID <= 0 {
		return nil, fmt.Errorf("%w: invalid product ID", ErrInvalidInput)
	}
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be greater than zero", ErrInvalidInput)
	}

	item := model.CartItem{
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
		Price:       req.Price,
		Quantity:    req.Quantity,
	}

	if err := s.repo.SaveItem(ctx, userID, item); err != nil {
		return nil, err
	}

	return s.repo.GetCart(ctx, userID)
}

func (s *cartServiceImpl) UpdateItemQuantity(ctx context.Context, userID int, productID int, req model.UpdateItemQuantityRequest) (*model.Cart, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	if productID <= 0 {
		return nil, fmt.Errorf("%w: invalid product ID", ErrInvalidInput)
	}
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be greater than zero", ErrInvalidInput)
	}

	if err := s.repo.UpdateItemQuantity(ctx, userID, productID, req.Quantity); err != nil {
		return nil, err
	}

	return s.repo.GetCart(ctx, userID)
}

func (s *cartServiceImpl) RemoveFromCart(ctx context.Context, userID int, productID int) (*model.Cart, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	if productID <= 0 {
		return nil, fmt.Errorf("%w: invalid product ID", ErrInvalidInput)
	}

	if err := s.repo.RemoveItem(ctx, userID, productID); err != nil {
		return nil, err
	}

	return s.repo.GetCart(ctx, userID)
}

func (s *cartServiceImpl) ClearUserCart(ctx context.Context, userID int) error {
	if userID <= 0 {
		return fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	return s.repo.ClearCart(ctx, userID)
}
