package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserService defines the interface for user business logic
type UserService interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, email, password string) (string, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// ProductService defines the interface for product business logic
type ProductService interface {
	Create(ctx context.Context, product *Product, userID uuid.UUID) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Product, error)
	GetAllByUser(ctx context.Context, userID uuid.UUID) ([]Product, error)
	Update(ctx context.Context, product *Product, userID uuid.UUID) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
} 