package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the generic interface for CRUD operations
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	GetAll(ctx context.Context) ([]T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines the interface for user-specific operations
type UserRepository interface {
	Repository[User]
	GetByEmail(ctx context.Context, email string) (*User, error)
}

// ProductRepository defines the interface for product-specific operations
type ProductRepository interface {
	Repository[Product]
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Product, error)
} 