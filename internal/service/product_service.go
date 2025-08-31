package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"products/internal/domain"
	"products/internal/repository"
)

// ProductService implements the product service interface
type ProductService struct {
	productRepo  *repository.ProductRepository
	cacheService *CacheService
}

// NewProductService creates a new product service
func NewProductService(productRepo *repository.ProductRepository, cacheService *CacheService) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		cacheService: cacheService,
	}
}

// Create creates a new product for a specific user
func (s *ProductService) Create(ctx context.Context, product *domain.Product, userID uuid.UUID) error {
	product.ID = uuid.New()
	product.UserID = userID
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	if err := s.productRepo.Create(ctx, product); err != nil {
		return err
	}

	s.invalidateUserCache(ctx, userID)

	return nil
}

// GetByID retrieves a product by ID, ensuring the user owns it
func (s *ProductService) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Product, error) {
	cacheKey := fmt.Sprintf("product:%s:%s", userID, id)
	var cachedProduct domain.Product
	if err := s.cacheService.Get(ctx, cacheKey, &cachedProduct); err == nil {
		return &cachedProduct, nil
	}

	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if product.UserID != userID {
		return nil, errors.New("unauthorized access to product")
	}

	s.cacheService.Set(ctx, cacheKey, product, 30*time.Minute)

	return product, nil
}

// GetAllByUser retrieves all products for a specific user
func (s *ProductService) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]domain.Product, error) {
	cacheKey := fmt.Sprintf("user_products:%s", userID)
	var cachedProducts []domain.Product
	if err := s.cacheService.Get(ctx, cacheKey, &cachedProducts); err == nil {
		return cachedProducts, nil
	}

	products, err := s.productRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.cacheService.Set(ctx, cacheKey, products, 15*time.Minute)

	return products, nil
}

// GetProductsWithFilters retrieves products with advanced filtering, sorting, and pagination
func (s *ProductService) GetProductsWithFilters(ctx context.Context, userID uuid.UUID, query domain.ProductQuery) (*domain.ProductListResponse, error) {
	cacheKey := s.generateQueryCacheKey(userID, query)

	var cachedResponse domain.ProductListResponse
	if err := s.cacheService.Get(ctx, cacheKey, &cachedResponse); err == nil {
		return &cachedResponse, nil
	}

	response, err := s.productRepo.GetProductsWithFilters(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	s.cacheService.Set(ctx, cacheKey, response, 5*time.Minute)

	return response, nil
}

// GetProductsWithCursor retrieves products with cursor-based pagination
func (s *ProductService) GetProductsWithCursor(ctx context.Context, userID uuid.UUID, query domain.ProductQueryCursor) (*domain.ProductListCursorResponse, error) {
	cacheKey := s.generateCursorQueryCacheKey(userID, query)

	var cachedResponse domain.ProductListCursorResponse
	if err := s.cacheService.Get(ctx, cacheKey, &cachedResponse); err == nil {
		return &cachedResponse, nil
	}

	response, err := s.productRepo.GetProductsWithCursor(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	s.cacheService.Set(ctx, cacheKey, response, 5*time.Minute)

	return response, nil
}

// Update updates a product, ensuring the user owns it
func (s *ProductService) Update(ctx context.Context, product *domain.Product, userID uuid.UUID) error {
	existingProduct, err := s.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		return err
	}

	if existingProduct.UserID != userID {
		return errors.New("unauthorized access to product")
	}

	if product.Name != "" {
		existingProduct.Name = product.Name
	}
	if product.Description != "" {
		existingProduct.Description = product.Description
	}
	if product.Price > 0 {
		existingProduct.Price = product.Price
	}
	if product.Stock >= 0 {
		existingProduct.Stock = product.Stock
	}

	existingProduct.UpdatedAt = time.Now()

	if err := s.productRepo.Update(ctx, existingProduct); err != nil {
		return err
	}

	s.invalidateUserCache(ctx, userID)

	return nil
}

// Delete deletes a product, ensuring the user owns it
func (s *ProductService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	existingProduct, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existingProduct.UserID != userID {
		return errors.New("unauthorized access to product")
	}

	if err := s.productRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.invalidateUserCache(ctx, userID)

	return nil
}

// GetProductStats retrieves product statistics for a user
func (s *ProductService) GetProductStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("user_stats:%s", userID)
	var cachedStats map[string]interface{}
	if err := s.cacheService.Get(ctx, cacheKey, &cachedStats); err == nil {
		return cachedStats, nil
	}

	stats, err := s.productRepo.GetProductStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.cacheService.Set(ctx, cacheKey, stats, 10*time.Minute)

	return stats, nil
}

// generateQueryCacheKey generates a cache key for filtered queries
func (s *ProductService) generateQueryCacheKey(userID uuid.UUID, query domain.ProductQuery) string {
	queryBytes, _ := json.Marshal(query)
	return fmt.Sprintf("user_products_filtered:%s:%s", userID, string(queryBytes))
}

// generateCursorQueryCacheKey generates a cache key for cursor-based queries
func (s *ProductService) generateCursorQueryCacheKey(userID uuid.UUID, query domain.ProductQueryCursor) string {
	queryBytes, _ := json.Marshal(query)
	return fmt.Sprintf("user_products_cursor:%s:%s", userID, string(queryBytes))
}

// invalidateUserCache invalidates all cache entries for a specific user
func (s *ProductService) invalidateUserCache(ctx context.Context, userID uuid.UUID) {
	s.cacheService.Delete(ctx, fmt.Sprintf("user_products:%s", userID))

	s.cacheService.Delete(ctx, fmt.Sprintf("user_stats:%s", userID))

	pattern := fmt.Sprintf("user_products_filtered:%s:*", userID)
	s.cacheService.DeletePattern(ctx, pattern)

	pattern = fmt.Sprintf("user_products_cursor:%s:*", userID)
	s.cacheService.DeletePattern(ctx, pattern)
}
