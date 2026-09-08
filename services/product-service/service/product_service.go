package service

import (
	"errors"
	"fmt"
	"strings"

	"product-service/model"
	"product-service/repository"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type ProductService interface {
	CreateProduct(req model.CreateProductRequest) (*model.Product, error)
	GetProduct(id int) (*model.Product, error)
	ListProducts() ([]model.Product, error)
	UpdateProduct(id int, req model.UpdateProductRequest) (*model.Product, error)
	DeleteProduct(id int) error
}

type productServiceImpl struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productServiceImpl{repo: repo}
}

func (s *productServiceImpl) CreateProduct(req model.CreateProductRequest) (*model.Product, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: product name is required", ErrInvalidInput)
	}

	if req.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be greater than zero", ErrInvalidInput)
	}

	if req.Stock < 0 {
		return nil, fmt.Errorf("%w: stock cannot be negative", ErrInvalidInput)
	}

	category := strings.TrimSpace(req.Category)
	if category == "" {
		return nil, fmt.Errorf("%w: product category is required", ErrInvalidInput)
	}

	product := &model.Product{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    category,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *productServiceImpl) GetProduct(id int) (*model.Product, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid product ID", ErrInvalidInput)
	}
	return s.repo.GetByID(id)
}

func (s *productServiceImpl) ListProducts() ([]model.Product, error) {
	return s.repo.GetAll()
}

func (s *productServiceImpl) UpdateProduct(id int, req model.UpdateProductRequest) (*model.Product, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid product ID", ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name != "" {
		existing.Name = name
	}

	if req.Description != "" {
		existing.Description = strings.TrimSpace(req.Description)
	}

	if req.Price > 0 {
		existing.Price = req.Price
	}

	if req.Stock >= 0 {
		existing.Stock = req.Stock
	}

	category := strings.TrimSpace(req.Category)
	if category != "" {
		existing.Category = category
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *productServiceImpl) DeleteProduct(id int) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid product ID", ErrInvalidInput)
	}
	return s.repo.Delete(id)
}
