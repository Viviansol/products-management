package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUser_TableName(t *testing.T) {
	user := &User{}
	if user.TableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", user.TableName())
	}
}

func TestProduct_TableName(t *testing.T) {
	product := &Product{}
	if product.TableName() != "products" {
		t.Errorf("Expected table name 'products', got '%s'", product.TableName())
	}
}

func TestUser_Creation(t *testing.T) {
	user := &User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", user.Name)
	}
}

func TestProduct_Creation(t *testing.T) {
	userID := uuid.New()
	product := &Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       29.99,
		Stock:       100,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if product.Name != "Test Product" {
		t.Errorf("Expected name 'Test Product', got '%s'", product.Name)
	}

	if product.Price != 29.99 {
		t.Errorf("Expected price 29.99, got %f", product.Price)
	}

	if product.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, product.UserID)
	}
}
