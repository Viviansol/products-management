package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"products/internal/domain"
	"products/internal/service"
	"products/cmd/api/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// validateUUID validates if the string is a valid UUID
func validateUUID(id string) (uuid.UUID, error) {
	if id == "" {
		return uuid.Nil, errors.New("ID is required")
	}
	
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.New("invalid ID format")
	}
	
	return parsedID, nil
}

// Create handles product creation with enhanced validation
func (h *ProductHandler) Create(c *gin.Context) {
	var req domain.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Sanitize inputs
	req.Name = validation.SanitizeInput(req.Name)
	req.Description = validation.SanitizeInput(req.Description)
	
	// Validate product name
	if err := validation.ValidateProductName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
		})
		return
	}
	
	// Validate description
	if err := validation.ValidateDescription(req.Description); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
		})
		return
	}
	
	// Validate price
	if err := validation.ValidatePrice(req.Price); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
		})
		return
	}
	
	// Validate stock
	if err := validation.ValidateStock(req.Stock); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
		})
		return
	}
	
	// Check for SQL injection patterns
	if validation.CheckSQLInjection(req.Name) || validation.CheckSQLInjection(req.Description) {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Security Error",
			Message: "Invalid input detected",
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	product := &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := h.productService.Create(c.Request.Context(), product, userID); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Creation Failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetByID handles retrieving a product by ID with enhanced validation
func (h *ProductHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	
	// Validate UUID format
	id, err := validateUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	product, err := h.productService.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Error:   "Not Found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetAllByUser handles retrieving all products for the authenticated user
func (h *ProductHandler) GetAllByUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	products, err := h.productService.GetAllByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to retrieve products",
		})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProductsWithFilters handles advanced product querying with filters, sorting, and pagination
func (h *ProductHandler) GetProductsWithFilters(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	// Parse query parameters
	query := domain.ProductQuery{
		Filter: domain.ProductFilter{},
		Sort:   []domain.SortField{},
		Pagination: domain.Pagination{
			Page:     1,
			PageSize: 20,
		},
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Pagination.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			query.Pagination.PageSize = pageSize
		}
	}

	// Parse filters
	if name := c.Query("name"); name != "" {
		query.Filter.Name = &name
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			query.Filter.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			query.Filter.MaxPrice = &maxPrice
		}
	}

	if minStockStr := c.Query("min_stock"); minStockStr != "" {
		if minStock, err := strconv.Atoi(minStockStr); err == nil {
			query.Filter.MinStock = &minStock
		}
	}

	if maxStockStr := c.Query("max_stock"); maxStockStr != "" {
		if maxStock, err := strconv.Atoi(maxStockStr); err == nil {
			query.Filter.MaxStock = &maxStock
		}
	}

	if createdFromStr := c.Query("created_from"); createdFromStr != "" {
		if createdFrom, err := time.Parse(time.RFC3339, createdFromStr); err == nil {
			query.Filter.CreatedFrom = &createdFrom
		}
	}

	if createdToStr := c.Query("created_to"); createdToStr != "" {
		if createdTo, err := time.Parse(time.RFC3339, createdToStr); err == nil {
			query.Filter.CreatedTo = &createdTo
		}
	}

	// Parse sorting
	if sortField := c.Query("sort_field"); sortField != "" {
		sortDirection := c.DefaultQuery("sort_direction", "asc")
		query.Sort = append(query.Sort, domain.SortField{
			Field:     sortField,
			Direction: sortDirection,
		})
	}

	response, err := h.productService.GetProductsWithFilters(c.Request.Context(), userID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to retrieve products",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProductsWithCursor handles cursor-based pagination
func (h *ProductHandler) GetProductsWithCursor(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	query := domain.ProductQueryCursor{
		Filter: domain.ProductFilter{},
		Sort:   []domain.SortField{},
		Pagination: domain.CursorPagination{
			PageSize: 20,
		},
	}

	// Parse cursor pagination
	if cursor := c.Query("cursor"); cursor != "" {
		query.Pagination.Cursor = &cursor
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			query.Pagination.PageSize = pageSize
		}
	}

	// Parse filters (same as above)
	if name := c.Query("name"); name != "" {
		query.Filter.Name = &name
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			query.Filter.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			query.Filter.MaxPrice = &maxPrice
		}
	}

	// Parse sorting
	if sortField := c.Query("sort_field"); sortField != "" {
		sortDirection := c.DefaultQuery("sort_direction", "asc")
		query.Sort = append(query.Sort, domain.SortField{
			Field:     sortField,
			Direction: sortDirection,
		})
	}

	response, err := h.productService.GetProductsWithCursor(c.Request.Context(), userID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to retrieve products",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProductStats retrieves product statistics for the authenticated user
func (h *ProductHandler) GetProductStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	stats, err := h.productService.GetProductStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to retrieve product statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Update handles product updates with enhanced validation
func (h *ProductHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	
	// Validate UUID format
	id, err := validateUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
		})
		return
	}

	var req domain.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	// Validate provided fields
	if req.Name != nil {
		*req.Name = validation.SanitizeInput(*req.Name)
		if err := validation.ValidateProductName(*req.Name); err != nil {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Validation Error",
				Message: "Name: " + err.Error(),
			})
			return
		}
		if validation.CheckSQLInjection(*req.Name) {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Security Error",
				Message: "Invalid name input detected",
			})
			return
		}
	}
	
	if req.Description != nil {
		*req.Description = validation.SanitizeInput(*req.Description)
		if err := validation.ValidateDescription(*req.Description); err != nil {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Validation Error",
				Message: "Description: " + err.Error(),
			})
			return
		}
		if validation.CheckSQLInjection(*req.Description) {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Security Error",
				Message: "Invalid description input detected",
			})
			return
		}
	}
	
	if req.Price != nil {
		if err := validation.ValidatePrice(*req.Price); err != nil {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Validation Error",
				Message: "Price: " + err.Error(),
			})
			return
		}
	}
	
	if req.Stock != nil {
		if err := validation.ValidateStock(*req.Stock); err != nil {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Validation Error",
				Message: "Stock: " + err.Error(),
			})
			return
		}
	}

	// Create product with only the fields to update
	product := &domain.Product{
		ID: id,
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}

	if err := h.productService.Update(c.Request.Context(), product, userID); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Update Failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

// Delete handles product deletion with enhanced validation
func (h *ProductHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	
	// Validate UUID format
	id, err := validateUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.productService.Delete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Deletion Failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
} 