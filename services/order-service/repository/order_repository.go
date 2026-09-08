package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"order-service/model"

	_ "github.com/lib/pq"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type OrderRepository interface {
	CreateOrderWithItems(ctx context.Context, order *model.Order, items []model.OrderItem) error
	GetOrderByID(ctx context.Context, id int) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error)
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) CreateOrderWithItems(ctx context.Context, order *model.Order, items []model.OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	orderQuery := `
		INSERT INTO orders (user_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, orderQuery, order.UserID, order.Status, order.TotalAmount).
		Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	stmt, err := tx.PrepareContext(ctx, itemQuery)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to prepare order item statement: %w", err)
	}
	defer stmt.Close()

	for i := range items {
		items[i].OrderID = order.ID
		err = stmt.QueryRowContext(ctx, order.ID, items[i].ProductID, items[i].Quantity, items[i].Price).
			Scan(&items[i].ID)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	order.Items = items
	return nil
}

func (r *PostgresOrderRepository) GetOrderByID(ctx context.Context, id int) (*model.Order, error) {
	orderQuery := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	order := &model.Order{}
	err := r.db.QueryRowContext(ctx, orderQuery, id).
		Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.CreatedAt, &order.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, order_id, product_id, quantity, price
		FROM order_items
		WHERE order_id = $1
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	order.Items = []model.OrderItem{}
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	ordersQuery := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY id DESC
	`
	rows, err := r.db.QueryContext(ctx, ordersQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []model.Order{}
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		o.Items = []model.OrderItem{}
		orders = append(orders, o)
	}

	// Fetch items for each order
	for i := range orders {
		itemsQuery := `
			SELECT id, order_id, product_id, quantity, price
			FROM order_items
			WHERE order_id = $1
			ORDER BY id ASC
		`
		itemRows, err := r.db.QueryContext(ctx, itemsQuery, orders[i].ID)
		if err != nil {
			return nil, err
		}
		for itemRows.Next() {
			var it model.OrderItem
			if err := itemRows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Quantity, &it.Price); err == nil {
				orders[i].Items = append(orders[i].Items, it)
			}
		}
		itemRows.Close()
	}

	return orders, nil
}

// In-memory mock repository for tests and local execution
type MockOrderRepository struct {
	mu         sync.RWMutex
	orders     map[int]*model.Order
	orderItems map[int][]model.OrderItem
	orderSeq   int
	itemSeq    int
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders:     make(map[int]*model.Order),
		orderItems: make(map[int][]model.OrderItem),
		orderSeq:   1,
		itemSeq:    1,
	}
}

func (m *MockOrderRepository) CreateOrderWithItems(ctx context.Context, order *model.Order, items []model.OrderItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	order.ID = m.orderSeq
	m.orderSeq++
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	savedItems := make([]model.OrderItem, len(items))
	for i, item := range items {
		item.ID = m.itemSeq
		m.itemSeq++
		item.OrderID = order.ID
		savedItems[i] = item
	}

	order.Items = savedItems

	cp := *order
	m.orders[order.ID] = &cp
	m.orderItems[order.ID] = savedItems
	return nil
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, id int) (*model.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	order, ok := m.orders[id]
	if !ok {
		return nil, ErrOrderNotFound
	}

	cp := *order
	cp.Items = m.orderItems[id]
	return &cp, nil
}

func (m *MockOrderRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := []model.Order{}
	for _, o := range m.orders {
		if o.UserID == userID {
			cp := *o
			cp.Items = m.orderItems[o.ID]
			result = append(result, cp)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID > result[j].ID
	})

	return result, nil
}
