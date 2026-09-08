package repository

import (
	"database/sql"
	"errors"
	"sort"
	"sync"
	"time"

	"product-service/model"

	_ "github.com/lib/pq"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductRepository interface {
	Create(p *model.Product) error
	GetByID(id int) (*model.Product, error)
	GetAll() ([]model.Product, error)
	Update(p *model.Product) error
	Delete(id int) error
}

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(p *model.Product) error {
	query := `
		INSERT INTO products (name, description, price, stock, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, p.Name, p.Description, p.Price, p.Stock, p.Category).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PostgresProductRepository) GetByID(id int) (*model.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	p := &model.Product{}
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PostgresProductRepository) GetAll() ([]model.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category, created_at, updated_at
		FROM products
		ORDER BY id ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []model.Product{}
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *PostgresProductRepository) Update(p *model.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4, category = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`
	err := r.db.QueryRow(query, p.Name, p.Description, p.Price, p.Stock, p.Category, p.ID).Scan(&p.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrProductNotFound
	}
	return err
}

func (r *PostgresProductRepository) Delete(id int) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

// In-memory Mock Product Repository for tests and standalone execution
type MockProductRepository struct {
	mu       sync.RWMutex
	products map[int]*model.Product
	idSeq    int
}

func NewMockProductRepository() *MockProductRepository {
	repo := &MockProductRepository{
		products: make(map[int]*model.Product),
		idSeq:    1,
	}
	repo.seedData()
	return repo
}

func (m *MockProductRepository) seedData() {
	seeds := []struct {
		Name        string
		Description string
		Price       float64
		Stock       int
		Category    string
	}{
		{"ProBook Laptop 15\"", "High-performance laptop with 16GB RAM and 512GB SSD", 1299.99, 25, "Electronics"},
		{"UltraPhone 14 Pro", "Flagship smartphone with triple camera system and OLED display", 999.00, 40, "Electronics"},
		{"NoiseCancelling Headphones", "Wireless over-ear headphones with active noise cancellation", 249.50, 60, "Audio"},
		{"Mechanical RGB Keyboard", "Tactile mechanical keyboard with customizable RGB backlighting", 89.99, 100, "Accessories"},
		{"Wireless Ergonomic Mouse", "Precision optical mouse with multi-device bluetooth pairing", 49.99, 120, "Accessories"},
		{"4K UltraHD Monitor 27\"", "IPS display panel with 144Hz refresh rate and HDR400", 399.99, 15, "Electronics"},
		{"SmartWatch Pro V2", "Fitness tracking smartwatch with heart rate & ECG sensors", 199.95, 50, "Wearables"},
		{"OctaTab 11 Tablet", "11-inch tablet with stylus support and all-day battery life", 499.00, 30, "Electronics"},
		{"Mirrorless 4K Camera", "Professional digital camera with 24MP sensor and 4K video recording", 849.99, 10, "Photography"},
		{"NextGen Gaming Console 1TB", "Ultra-fast SSD gaming console supporting up to 120 FPS output", 499.99, 20, "Gaming"},
	}

	for _, s := range seeds {
		p := &model.Product{
			Name:        s.Name,
			Description: s.Description,
			Price:       s.Price,
			Stock:       s.Stock,
			Category:    s.Category,
		}
		_ = m.Create(p)
	}
}

func (m *MockProductRepository) Create(p *model.Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p.ID = m.idSeq
	m.idSeq++
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	// Clone to store
	cp := *p
	m.products[p.ID] = &cp
	return nil
}

func (m *MockProductRepository) GetByID(id int) (*model.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.products[id]
	if !ok {
		return nil, ErrProductNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *MockProductRepository) GetAll() ([]model.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]model.Product, 0, len(m.products))
	for _, p := range m.products {
		list = append(list, *p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list, nil
}

func (m *MockProductRepository) Update(p *model.Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.products[p.ID]
	if !ok {
		return ErrProductNotFound
	}

	existing.Name = p.Name
	existing.Description = p.Description
	existing.Price = p.Price
	existing.Stock = p.Stock
	existing.Category = p.Category
	existing.UpdatedAt = time.Now()

	p.UpdatedAt = existing.UpdatedAt
	p.CreatedAt = existing.CreatedAt
	return nil
}

func (m *MockProductRepository) Delete(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.products[id]; !ok {
		return ErrProductNotFound
	}
	delete(m.products, id)
	return nil
}
