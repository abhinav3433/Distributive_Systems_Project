package repository

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"payment-service/model"

	_ "github.com/lib/pq"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

type PaymentRepository interface {
	Create(ctx context.Context, p *model.Payment) error
	GetByOrderID(ctx context.Context, orderID int) (*model.Payment, error)
}

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Create(ctx context.Context, p *model.Payment) error {
	query := `
		INSERT INTO payments (order_id, user_id, amount, status, payment_method, transaction_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query, p.OrderID, p.UserID, p.Amount, p.Status, p.PaymentMethod, p.TransactionID).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID int) (*model.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, status, payment_method, transaction_id, created_at, updated_at
		FROM payments
		WHERE order_id = $1
		ORDER BY id DESC
		LIMIT 1
	`
	p := &model.Payment{}
	err := r.db.QueryRowContext(ctx, query, orderID).
		Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Status, &p.PaymentMethod, &p.TransactionID, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

// In-memory Mock Payment Repository
type MockPaymentRepository struct {
	mu       sync.RWMutex
	payments map[int]*model.Payment
	seq      int
}

func NewMockPaymentRepository() *MockPaymentRepository {
	return &MockPaymentRepository{
		payments: make(map[int]*model.Payment),
		seq:      1,
	}
}

func (m *MockPaymentRepository) Create(ctx context.Context, p *model.Payment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p.ID = m.seq
	m.seq++
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	cp := *p
	m.payments[p.OrderID] = &cp
	return nil
}

func (m *MockPaymentRepository) GetByOrderID(ctx context.Context, orderID int) (*model.Payment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.payments[orderID]
	if !ok {
		return nil, ErrPaymentNotFound
	}
	cp := *p
	return &cp, nil
}
