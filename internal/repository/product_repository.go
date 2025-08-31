package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"products/internal/domain"
)

// ProductRepository implements the product repository interface
type ProductRepository struct {
	*GenericRepository[domain.Product]
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		GenericRepository: NewGenericRepository[domain.Product](db),
		db:                db,
	}
}

// GetByUserID retrieves all products for a specific user
func (r *ProductRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&products).Error
	return products, err
}

// GetByID retrieves a product by ID with user information
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	return &product, nil
}

// GetProductsWithFilters retrieves products with advanced filtering, sorting, and pagination
func (r *ProductRepository) GetProductsWithFilters(ctx context.Context, userID uuid.UUID, query domain.ProductQuery) (*domain.ProductListResponse, error) {
	var products []domain.Product
	var total int64

	dbQuery := r.db.WithContext(ctx).Where("user_id = ?", userID)

	dbQuery = r.applyFilters(dbQuery, query.Filter)

	if err := dbQuery.Model(&domain.Product{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	dbQuery = r.applySorting(dbQuery, query.Sort)

	offset := (query.Pagination.Page - 1) * query.Pagination.PageSize
	dbQuery = dbQuery.Offset(offset).Limit(query.Pagination.PageSize)

	if err := dbQuery.Preload("User").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	totalPages := int((total + int64(query.Pagination.PageSize) - 1) / int64(query.Pagination.PageSize))
	hasNext := query.Pagination.Page < totalPages
	hasPrev := query.Pagination.Page > 1

	return &domain.ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       query.Pagination.Page,
		PageSize:   query.Pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// GetProductsWithCursor retrieves products with cursor-based pagination
func (r *ProductRepository) GetProductsWithCursor(ctx context.Context, userID uuid.UUID, query domain.ProductQueryCursor) (*domain.ProductListCursorResponse, error) {
	var products []domain.Product

	dbQuery := r.db.WithContext(ctx).Where("user_id = ?", userID)

	dbQuery = r.applyFilters(dbQuery, query.Filter)

	dbQuery = r.applySorting(dbQuery, query.Sort)

	if query.Pagination.Cursor != nil {
		cursor, err := uuid.Parse(*query.Pagination.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}

		dbQuery = dbQuery.Where("id > ?", cursor)
	}

	limit := query.Pagination.PageSize + 1
	if err := dbQuery.Preload("User").Limit(limit).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	hasNext := len(products) > query.Pagination.PageSize
	if hasNext {
		products = products[:query.Pagination.PageSize]
	}

	var nextCursor, prevCursor *string
	if len(products) > 0 {
		lastID := products[len(products)-1].ID.String()
		nextCursor = &lastID

		if query.Pagination.Cursor != nil {
			firstID := products[0].ID.String()
			prevCursor = &firstID
		}
	}

	return &domain.ProductListCursorResponse{
		Products:   products,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    query.Pagination.Cursor != nil,
	}, nil
}

// applyFilters applies filters to the database query
func (r *ProductRepository) applyFilters(dbQuery *gorm.DB, filter domain.ProductFilter) *gorm.DB {
	if filter.Name != nil && *filter.Name != "" {
		dbQuery = dbQuery.Where("LOWER(name) LIKE LOWER(?)", "%"+*filter.Name+"%")
	}

	if filter.MinPrice != nil {
		dbQuery = dbQuery.Where("price >= ?", *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		dbQuery = dbQuery.Where("price <= ?", *filter.MaxPrice)
	}

	if filter.MinStock != nil {
		dbQuery = dbQuery.Where("stock >= ?", *filter.MinStock)
	}

	if filter.MaxStock != nil {
		dbQuery = dbQuery.Where("stock <= ?", *filter.MaxStock)
	}

	if filter.CreatedFrom != nil {
		dbQuery = dbQuery.Where("created_at >= ?", *filter.CreatedFrom)
	}

	if filter.CreatedTo != nil {
		dbQuery = dbQuery.Where("created_at <= ?", *filter.CreatedTo)
	}

	if filter.UpdatedFrom != nil {
		dbQuery = dbQuery.Where("updated_at >= ?", *filter.UpdatedFrom)
	}

	if filter.UpdatedTo != nil {
		dbQuery = dbQuery.Where("updated_at <= ?", *filter.UpdatedTo)
	}

	return dbQuery
}

// applySorting applies sorting to the database query
func (r *ProductRepository) applySorting(dbQuery *gorm.DB, sortFields []domain.SortField) *gorm.DB {
	if len(sortFields) == 0 {
		// Default sorting by created_at desc
		return dbQuery.Order("created_at DESC")
	}

	for _, sortField := range sortFields {
		field := sortField.Field
		direction := strings.ToUpper(sortField.Direction)

		validFields := map[string]bool{
			"name":       true,
			"price":      true,
			"stock":      true,
			"created_at": true,
			"updated_at": true,
		}

		if !validFields[field] {
			continue
		}

		if direction != "ASC" && direction != "DESC" {
			direction = "ASC"
		}

		dbQuery = dbQuery.Order(fmt.Sprintf("%s %s", field, direction))
	}

	return dbQuery
}

// GetProductStats retrieves product statistics for a user
func (r *ProductRepository) GetProductStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	var stats struct {
		TotalProducts int64   `json:"total_products"`
		TotalValue    float64 `json:"total_value"`
		AvgPrice      float64 `json:"avg_price"`
		LowStock      int64   `json:"low_stock"`
		OutOfStock    int64   `json:"out_of_stock"`
	}

	err := r.db.WithContext(ctx).
		Model(&domain.Product{}).
		Where("user_id = ?", userID).
		Select(`
			COUNT(*) as total_products,
			COALESCE(SUM(price * stock), 0) as total_value,
			COALESCE(AVG(price), 0) as avg_price,
			COUNT(CASE WHEN stock < 10 THEN 1 END) as low_stock,
			COUNT(CASE WHEN stock = 0 THEN 1 END) as out_of_stock
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get product stats: %w", err)
	}

	return map[string]interface{}{
		"total_products": stats.TotalProducts,
		"total_value":    stats.TotalValue,
		"avg_price":      stats.AvgPrice,
		"low_stock":      stats.LowStock,
		"out_of_stock":   stats.OutOfStock,
	}, nil
}
