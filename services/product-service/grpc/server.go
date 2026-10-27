package grpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"product-service/model"
	"product-service/repository"
	"product-service/service"
	productv1 "proto/product/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductGRPCServer struct {
	productv1.UnimplementedProductServiceServer
	productSvc service.ProductService
}

func NewProductGRPCServer(productSvc service.ProductService) *ProductGRPCServer {
	return &ProductGRPCServer{
		productSvc: productSvc,
	}
}

func toProtoProduct(p *model.Product) *productv1.Product {
	if p == nil {
		return nil
	}
	return &productv1.Product{
		Id:          int64(p.ID),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       int32(p.Stock),
		Category:    p.Category,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *ProductGRPCServer) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	products, err := s.productSvc.ListProducts()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list products: "+err.Error())
	}

	categoryFilter := ""
	if req != nil {
		categoryFilter = strings.TrimSpace(strings.ToLower(req.GetCategory()))
	}

	var protoProducts []*productv1.Product
	for _, p := range products {
		if categoryFilter != "" && categoryFilter != "all" && strings.ToLower(p.Category) != categoryFilter {
			continue
		}
		prodCopy := p
		protoProducts = append(protoProducts, toProtoProduct(&prodCopy))
	}

	return &productv1.ListProductsResponse{
		Products: protoProducts,
	}, nil
}

func (s *ProductGRPCServer) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid product id is required")
	}

	p, err := s.productSvc.GetProduct(int(req.GetId()))
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, "failed to get product: "+err.Error())
	}

	return &productv1.GetProductResponse{
		Product: toProtoProduct(p),
	}, nil
}

func (s *ProductGRPCServer) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	p, err := s.productSvc.CreateProduct(model.CreateProductRequest{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Price:       req.GetPrice(),
		Stock:       int(req.GetStock()),
		Category:    req.GetCategory(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create product: "+err.Error())
	}

	return &productv1.CreateProductResponse{
		Product: toProtoProduct(p),
	}, nil
}

func (s *ProductGRPCServer) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid product id is required")
	}

	p, err := s.productSvc.UpdateProduct(int(req.GetId()), model.UpdateProductRequest{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Price:       req.GetPrice(),
		Stock:       int(req.GetStock()),
		Category:    req.GetCategory(),
	})
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update product: "+err.Error())
	}

	return &productv1.UpdateProductResponse{
		Product: toProtoProduct(p),
	}, nil
}

func (s *ProductGRPCServer) DeleteProduct(ctx context.Context, req *productv1.DeleteProductRequest) (*productv1.DeleteProductResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid product id is required")
	}

	err := s.productSvc.DeleteProduct(int(req.GetId()))
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete product: "+err.Error())
	}

	return &productv1.DeleteProductResponse{
		Success: true,
	}, nil
}
