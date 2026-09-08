package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"order-service/events"
	"order-service/messaging"
	"order-service/model"
	"order-service/repository"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type ProductDTO struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type OrderService interface {
	CreateOrder(ctx context.Context, userID int, req model.CreateOrderRequest) (*model.Order, error)
	GetOrderByID(ctx context.Context, id int) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error)
}

type orderServiceImpl struct {
	repo              repository.OrderRepository
	publisher         messaging.EventPublisher
	productServiceURL string
	httpClient        *http.Client
}

func NewOrderService(repo repository.OrderRepository, publisher messaging.EventPublisher, productServiceURL string) OrderService {
	return &orderServiceImpl{
		repo:              repo,
		publisher:         publisher,
		productServiceURL: productServiceURL,
		httpClient:        &http.Client{Timeout: 3 * time.Second},
	}
}

func (s *orderServiceImpl) fetchProductInfo(productID int) (*ProductDTO, error) {
	if s.productServiceURL == "" {
		return nil, errors.New("product service URL not configured")
	}

	url := fmt.Sprintf("%s/products/%d", s.productServiceURL, productID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var p ProductDTO
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}

	return &p, nil
}

func (s *orderServiceImpl) CreateOrder(ctx context.Context, userID int, req model.CreateOrderRequest) (*model.Order, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: order must contain at least one item", ErrInvalidInput)
	}

	orderItems := make([]model.OrderItem, len(req.Items))
	var totalAmount float64

	for i, it := range req.Items {
		if it.ProductID <= 0 {
			return nil, fmt.Errorf("%w: invalid product ID %d", ErrInvalidInput, it.ProductID)
		}
		if it.Quantity <= 0 {
			return nil, fmt.Errorf("%w: invalid quantity %d for product %d", ErrInvalidInput, it.Quantity, it.ProductID)
		}

		price := it.Price
		name := it.ProductName

		// Attempt to fetch fresh product information from Product Service
		prod, err := s.fetchProductInfo(it.ProductID)
		if err == nil && prod != nil && prod.Price > 0 {
			price = prod.Price
			if name == "" {
				name = prod.Name
			}
		}

		if price <= 0 {
			return nil, fmt.Errorf("%w: price for product %d could not be determined or is invalid", ErrInvalidInput, it.ProductID)
		}

		orderItems[i] = model.OrderItem{
			ProductID:   it.ProductID,
			ProductName: name,
			Quantity:    it.Quantity,
			Price:       price,
		}

		totalAmount += price * float64(it.Quantity)
	}

	order := &model.Order{
		UserID:          userID,
		Status:          model.StatusPending,
		TotalAmount:     totalAmount,
		ShippingAddress: req.ShippingAddress,
	}

	// 1. Save the order and order items in PostgreSQL within transaction
	if err := s.repo.CreateOrderWithItems(ctx, order, orderItems); err != nil {
		return nil, err
	}

	log.Printf("[ORDER] Order created (ID: %d, UserID: %d, Total: $%.2f)", order.ID, order.UserID, order.TotalAmount)

	// 2. Publish OrderCreated event to RabbitMQ
	if s.publisher != nil {
		event := events.OrderCreatedEvent{
			EventType: "OrderCreated",
			OrderID:   order.ID,
			UserID:    order.UserID,
			Amount:    order.TotalAmount,
		}
		if err := s.publisher.PublishOrderCreated(ctx, event); err != nil {
			log.Printf("Warning: Failed to publish OrderCreated event: %v", err)
		}
	}

	return order, nil
}

func (s *orderServiceImpl) GetOrderByID(ctx context.Context, id int) (*model.Order, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid order ID", ErrInvalidInput)
	}
	return s.repo.GetOrderByID(ctx, id)
}

func (s *orderServiceImpl) GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	return s.repo.GetOrdersByUserID(ctx, userID)
}
