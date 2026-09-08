package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"cart-service/model"

	"github.com/redis/go-redis/v9"
)

var (
	ErrItemNotFound = errors.New("cart item not found")
	ErrCartEmpty    = errors.New("cart is empty")
)

type CartRepository interface {
	GetCart(ctx context.Context, userID int) (*model.Cart, error)
	SaveItem(ctx context.Context, userID int, item model.CartItem) error
	UpdateItemQuantity(ctx context.Context, userID int, productID int, quantity int) error
	RemoveItem(ctx context.Context, userID int, productID int) error
	ClearCart(ctx context.Context, userID int) error
}

type RedisCartRepository struct {
	client *redis.Client
}

func NewRedisCartRepository(client *redis.Client) *RedisCartRepository {
	return &RedisCartRepository{client: client}
}

func getKey(userID int) string {
	return fmt.Sprintf("cart:%d", userID)
}

func (r *RedisCartRepository) GetCart(ctx context.Context, userID int) (*model.Cart, error) {
	key := getKey(userID)
	res, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis query error: %w", err)
	}

	cart := &model.Cart{
		UserID: userID,
		Items:  []model.CartItem{},
	}

	for _, val := range res {
		var item model.CartItem
		if err := json.Unmarshal([]byte(val), &item); err == nil {
			cart.Items = append(cart.Items, item)
			cart.TotalPrice += item.Price * float64(item.Quantity)
		}
	}

	return cart, nil
}

func (r *RedisCartRepository) SaveItem(ctx context.Context, userID int, item model.CartItem) error {
	key := getKey(userID)
	field := strconv.Itoa(item.ProductID)

	// Check if item already exists in hash to accumulate quantity
	existingVal, err := r.client.HGet(ctx, key, field).Result()
	if err == nil && existingVal != "" {
		var existingItem model.CartItem
		if json.Unmarshal([]byte(existingVal), &existingItem) == nil {
			item.Quantity += existingItem.Quantity
			if item.ProductName == "" {
				item.ProductName = existingItem.ProductName
			}
			if item.Price <= 0 {
				item.Price = existingItem.Price
			}
		}
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cart item: %w", err)
	}

	return r.client.HSet(ctx, key, field, data).Err()
}

func (r *RedisCartRepository) UpdateItemQuantity(ctx context.Context, userID int, productID int, quantity int) error {
	key := getKey(userID)
	field := strconv.Itoa(productID)

	existingVal, err := r.client.HGet(ctx, key, field).Result()
	if err == redis.Nil || existingVal == "" {
		return ErrItemNotFound
	}
	if err != nil {
		return err
	}

	var item model.CartItem
	if err := json.Unmarshal([]byte(existingVal), &item); err != nil {
		return err
	}

	item.Quantity = quantity
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return r.client.HSet(ctx, key, field, data).Err()
}

func (r *RedisCartRepository) RemoveItem(ctx context.Context, userID int, productID int) error {
	key := getKey(userID)
	field := strconv.Itoa(productID)

	res, err := r.client.HDel(ctx, key, field).Result()
	if err != nil {
		return err
	}
	if res == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *RedisCartRepository) ClearCart(ctx context.Context, userID int) error {
	key := getKey(userID)
	return r.client.Del(ctx, key).Err()
}

// In-memory Mock Repository for unit testing or fallback
type MockCartRepository struct {
	mu    sync.RWMutex
	carts map[int]map[int]model.CartItem
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{
		carts: make(map[int]map[int]model.CartItem),
	}
}

func (m *MockCartRepository) GetCart(ctx context.Context, userID int) (*model.Cart, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userCart, ok := m.carts[userID]
	cart := &model.Cart{
		UserID: userID,
		Items:  []model.CartItem{},
	}
	if !ok {
		return cart, nil
	}

	for _, item := range userCart {
		cart.Items = append(cart.Items, item)
		cart.TotalPrice += item.Price * float64(item.Quantity)
	}

	return cart, nil
}

func (m *MockCartRepository) SaveItem(ctx context.Context, userID int, item model.CartItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userCart, ok := m.carts[userID]
	if !ok {
		userCart = make(map[int]model.CartItem)
		m.carts[userID] = userCart
	}

	existing, exists := userCart[item.ProductID]
	if exists {
		item.Quantity += existing.Quantity
		if item.ProductName == "" {
			item.ProductName = existing.ProductName
		}
		if item.Price <= 0 {
			item.Price = existing.Price
		}
	}

	userCart[item.ProductID] = item
	return nil
}

func (m *MockCartRepository) UpdateItemQuantity(ctx context.Context, userID int, productID int, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userCart, ok := m.carts[userID]
	if !ok {
		return ErrItemNotFound
	}

	existing, exists := userCart[productID]
	if !exists {
		return ErrItemNotFound
	}

	existing.Quantity = quantity
	userCart[productID] = existing
	return nil
}

func (m *MockCartRepository) RemoveItem(ctx context.Context, userID int, productID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userCart, ok := m.carts[userID]
	if !ok {
		return ErrItemNotFound
	}

	if _, exists := userCart[productID]; !exists {
		return ErrItemNotFound
	}

	delete(userCart, productID)
	return nil
}

func (m *MockCartRepository) ClearCart(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.carts, userID)
	return nil
}
