package domain

import (
	"time"
)

// ProductFilter represents filters for product queries
type ProductFilter struct {
	Name        *string    `json:"name" form:"name"`
	MinPrice    *float64   `json:"min_price" form:"min_price"`
	MaxPrice    *float64   `json:"max_price" form:"max_price"`
	MinStock    *int       `json:"min_stock" form:"min_stock"`
	MaxStock    *int       `json:"max_stock" form:"max_stock"`
	CreatedFrom *time.Time `json:"created_from" form:"created_from"`
	CreatedTo   *time.Time `json:"created_to" form:"created_to"`
	UpdatedFrom *time.Time `json:"updated_from" form:"updated_from"`
	UpdatedTo   *time.Time `json:"updated_to" form:"updated_to"`
}

// SortField represents a field to sort by
type SortField struct {
	Field     string `json:"field" form:"field"`
	Direction string `json:"direction" form:"direction"` // "asc" or "desc"
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page" form:"page" binding:"min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100"`
}

// CursorPagination represents cursor-based pagination
type CursorPagination struct {
	Cursor   *string `json:"cursor" form:"cursor"`
	PageSize int     `json:"page_size" form:"page_size" binding:"min=1,max=100"`
}

// ProductQuery represents a complete product query with filters, sorting, and pagination
type ProductQuery struct {
	Filter     ProductFilter `json:"filter"`
	Sort       []SortField   `json:"sort"`
	Pagination Pagination    `json:"pagination"`
}

// ProductQueryCursor represents a cursor-based product query
type ProductQueryCursor struct {
	Filter     ProductFilter     `json:"filter"`
	Sort       []SortField       `json:"sort"`
	Pagination CursorPagination `json:"pagination"`
}

// ProductListResponse represents a paginated list of products
type ProductListResponse struct {
	Products   []Product `json:"products"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
	HasNext    bool      `json:"has_next"`
	HasPrev    bool      `json:"has_prev"`
}

// ProductListCursorResponse represents a cursor-based list of products
type ProductListCursorResponse struct {
	Products []Product `json:"products"`
	NextCursor *string `json:"next_cursor,omitempty"`
	PrevCursor *string `json:"prev_cursor,omitempty"`
	HasNext    bool    `json:"has_next"`
	HasPrev    bool    `json:"has_prev"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// SessionInfo represents session information
type SessionInfo struct {
	SessionID   string    `json:"session_id"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	IsActive    bool      `json:"is_active"`
}

// UserSessionsResponse represents user sessions information
type UserSessionsResponse struct {
	ActiveSessions []SessionInfo `json:"active_sessions"`
	TotalSessions  int64         `json:"total_sessions"`
}
