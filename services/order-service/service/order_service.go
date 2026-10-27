package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"order-service/events"
	"order-service/messaging"
	"order-service/model"
	"order-service/repository"
	productv1 "proto/product/v1"
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
	productGRPCClient productv1.ProductServiceClient
}

func NewOrderService(repo repository.OrderRepository, publisher messaging.EventPublisher, productServiceURL string) OrderService {
	return NewOrderServiceWithGRPC(repo, publisher, productServiceURL, nil)
}

func NewOrderServiceWithGRPC(repo repository.OrderRepository, publisher messaging.EventPublisher, productServiceURL string, productGRPCClient productv1.ProductServiceClient) OrderService {
	return &orderServiceImpl{
		repo:              repo,
		publisher:         publisher,
		productServiceURL: productServiceURL,
		httpClient:        &http.Client{Timeout: 3 * time.Second},
		productGRPCClient: productGRPCClient,
	}
}

func (s *orderServiceImpl) fetchProductInfo(ctx context.Context, productID int) (*ProductDTO, error) {
	// 1. Try gRPC first if client is configured
	if s.productGRPCClient != nil {
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		resp, err := s.productGRPCClient.GetProduct(reqCtx, &productv1.GetProductRequest{Id: int64(productID)})
		if err == nil && resp.GetProduct() != nil {
			p := resp.GetProduct()
			log.Printf("[ORDER-SERVICE] Successfully retrieved product %d via gRPC", productID)
			return &ProductDTO{
				ID:    int(p.GetId()),
				Name:  p.GetName(),
				Price: p.GetPrice(),
				Stock: int(p.GetStock()),
			}, nil
		}
		log.Printf("[ORDER-SERVICE] gRPC product lookup failed (%v), attempting HTTP fallback...", err)
	}

	// 2. Fall back to HTTP if gRPC is unavailable or unconfigured
	if s.productServiceURL == "" {
		return nil, errors.New("neither gRPC nor HTTP product service URL configured")
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

func validateShippingAddress(addr *model.ShippingAddress) error {
	if addr == nil {
		return nil
	}

	addr.RecipientName = strings.TrimSpace(addr.RecipientName)
	addr.PhoneNumber = strings.TrimSpace(addr.PhoneNumber)
	addr.AddressLine1 = strings.TrimSpace(addr.AddressLine1)
	addr.AddressLine2 = strings.TrimSpace(addr.AddressLine2)
	addr.City = strings.TrimSpace(addr.City)
	addr.State = strings.TrimSpace(addr.State)
	addr.PostalCode = strings.TrimSpace(addr.PostalCode)
	addr.Country = strings.TrimSpace(addr.Country)

	if addr.RecipientName == "" {
		return fmt.Errorf("%w: recipient name is required", ErrInvalidInput)
	}
	if len(addr.RecipientName) > 255 {
		return fmt.Errorf("%w: recipient name exceeds maximum length of 255 characters", ErrInvalidInput)
	}

	if addr.PhoneNumber == "" {
		return fmt.Errorf("%w: contact phone number is required", ErrInvalidInput)
	}
	if len(addr.PhoneNumber) > 50 {
		return fmt.Errorf("%w: contact phone number exceeds maximum length of 50 characters", ErrInvalidInput)
	}

	if addr.AddressLine1 == "" {
		return fmt.Errorf("%w: address line 1 is required", ErrInvalidInput)
	}
	if len(addr.AddressLine1) > 255 {
		return fmt.Errorf("%w: address line 1 exceeds maximum length of 255 characters", ErrInvalidInput)
	}

	if len(addr.AddressLine2) > 255 {
		return fmt.Errorf("%w: address line 2 exceeds maximum length of 255 characters", ErrInvalidInput)
	}

	if addr.City == "" {
		return fmt.Errorf("%w: city is required", ErrInvalidInput)
	}
	if len(addr.City) > 100 {
		return fmt.Errorf("%w: city exceeds maximum length of 100 characters", ErrInvalidInput)
	}

	if addr.State == "" {
		return fmt.Errorf("%w: state or province is required", ErrInvalidInput)
	}
	if len(addr.State) > 100 {
		return fmt.Errorf("%w: state or province exceeds maximum length of 100 characters", ErrInvalidInput)
	}

	if addr.PostalCode == "" {
		return fmt.Errorf("%w: postal code is required", ErrInvalidInput)
	}
	if len(addr.PostalCode) > 50 {
		return fmt.Errorf("%w: postal code exceeds maximum length of 50 characters", ErrInvalidInput)
	}

	if addr.Country == "" {
		return fmt.Errorf("%w: country is required", ErrInvalidInput)
	}
	if len(addr.Country) > 100 {
		return fmt.Errorf("%w: country exceeds maximum length of 100 characters", ErrInvalidInput)
	}

	return nil
}

func (s *orderServiceImpl) CreateOrder(ctx context.Context, userID int, req model.CreateOrderRequest) (*model.Order, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: order must contain at least one item", ErrInvalidInput)
	}

	// Validate shipping address snapshot if provided
	if req.Address != nil {
		if err := validateShippingAddress(req.Address); err != nil {
			return nil, err
		}
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
		prod, err := s.fetchProductInfo(ctx, it.ProductID)
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

	shippingSummary := req.ShippingAddress
	if req.Address != nil && shippingSummary == "" {
		shippingSummary = req.Address.FormatSummary()
	}

	order := &model.Order{
		UserID:          userID,
		Status:          model.StatusPending,
		TotalAmount:     totalAmount,
		ShippingAddress: shippingSummary,
		Address:         req.Address,
	}

	// 1. Save the order, items, and immutable shipping address in PostgreSQL within single transaction
	if err := s.repo.CreateOrderWithItems(ctx, order, orderItems); err != nil {
		return nil, err
	}

	log.Printf("[ORDER] Order created (ID: %d, UserID: %d, Total: $%.2f)", order.ID, order.UserID, order.TotalAmount)

	// 2. Publish OrderCreated event to RabbitMQ only after successful database commit
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
