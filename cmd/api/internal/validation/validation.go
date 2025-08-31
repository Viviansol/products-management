package validation

import (
	"errors"
	"regexp"
	"strings"
)

// Validation constants
const (
	MinPasswordLength    = 8
	MaxPasswordLength    = 128
	MinNameLength        = 2
	MaxNameLength        = 100
	MaxEmailLength       = 254
	MinProductNameLength = 2
	MaxProductNameLength = 200
	MaxDescriptionLength = 1000
	MinPrice            = 0.01
	MaxPrice            = 999999.99
	MinStock            = 0
	MaxStock            = 999999
)

// Validation regex patterns
var (
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	passwordRegex    = regexp.MustCompile(`^[A-Za-z\d@$!%*?&]{8,}$`)
	nameRegex        = regexp.MustCompile(`^[a-zA-Z\s\-'\.]+$`)
	productNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,!?()&]+$`)
	descriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,!?()&@#$%*+=:;'"<>[\]{}|\\/~]+$`)
)

// ValidateEmail validates email format and length
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	
	if email == "" {
		return errors.New("email is required")
	}
	
	if len(email) > MaxEmailLength {
		return errors.New("email is too long")
	}
	
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	
	return nil
}

// ValidatePassword validates password strength and length
func ValidatePassword(password string) error {
	password = strings.TrimSpace(password)
	
	if password == "" {
		return errors.New("password is required")
	}
	
	if len(password) < MinPasswordLength {
		return errors.New("password must be at least 8 characters long")
	}
	
	if len(password) > MaxPasswordLength {
		return errors.New("password is too long")
	}
	
	// Check for at least one lowercase letter
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return errors.New("password must contain at least one lowercase letter")
	}
	
	// Check for at least one uppercase letter
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return errors.New("password must contain at least one uppercase letter")
	}
	
	// Check for at least one number
	if !strings.ContainsAny(password, "0123456789") {
		return errors.New("password must contain at least one number")
	}
	
	// Check for at least one special character
	if !strings.ContainsAny(password, "@$!%*?&") {
		return errors.New("password must contain at least one special character (@$!%*?&)")
	}
	
	// Check for valid characters only
	if !passwordRegex.MatchString(password) {
		return errors.New("password contains invalid characters. Only letters, numbers, and @$!%*?& are allowed")
	}
	
	return nil
}

// ValidateName validates name format and length
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	
	if name == "" {
		return errors.New("name is required")
	}
	
	if len(name) < MinNameLength {
		return errors.New("name must be at least 2 characters long")
	}
	
	if len(name) > MaxNameLength {
		return errors.New("name is too long")
	}
	
	if !nameRegex.MatchString(name) {
		return errors.New("name contains invalid characters")
	}
	
	return nil
}

// ValidateProductName validates product name format and length
func ValidateProductName(name string) error {
	name = strings.TrimSpace(name)
	
	if name == "" {
		return errors.New("product name is required")
	}
	
	if len(name) < MinProductNameLength {
		return errors.New("product name must be at least 2 characters long")
	}
	
	if len(name) > MaxProductNameLength {
		return errors.New("product name is too long")
	}
	
	if !productNameRegex.MatchString(name) {
		return errors.New("product name contains invalid characters")
	}
	
	return nil
}

// ValidateDescription validates product description format and length
func ValidateDescription(description string) error {
	if description == "" {
		return nil // Description is optional
	}
	
	description = strings.TrimSpace(description)
	
	if len(description) > MaxDescriptionLength {
		return errors.New("description is too long")
	}
	
	if !descriptionRegex.MatchString(description) {
		return errors.New("description contains invalid characters")
	}
	
	return nil
}

// ValidatePrice validates product price range
func ValidatePrice(price float64) error {
	if price < MinPrice {
		return errors.New("price must be greater than 0")
	}
	
	if price > MaxPrice {
		return errors.New("price is too high")
	}
	
	return nil
}

// ValidateStock validates product stock range
func ValidateStock(stock int) error {
	if stock < MinStock {
		return errors.New("stock cannot be negative")
	}
	
	if stock > MaxStock {
		return errors.New("stock value is too high")
	}
	
	return nil
}

// SanitizeInput removes potentially dangerous characters
func SanitizeInput(input string) string {
	// Remove null bytes and control characters
	input = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, input)
	
	// Trim whitespace
	return strings.TrimSpace(input)
}

// CheckSQLInjection checks for common SQL injection patterns
func CheckSQLInjection(input string) bool {
	lowerInput := strings.ToLower(input)
	dangerousPatterns := []string{
		"union", "select", "insert", "update", "delete", "drop", "create",
		"alter", "exec", "execute", "script", "javascript", "vbscript",
		"<script", "javascript:", "onload", "onerror", "onclick",
	}
	
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	return false
}
