package handler

import (
	"net/http"
	"strings"

	"products/internal/domain"
	"products/internal/service"
	"products/cmd/api/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register handles user registration with enhanced validation
func (h *UserHandler) Register(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Sanitize inputs
	req.Email = validation.SanitizeInput(req.Email)
	req.Name = validation.SanitizeInput(req.Name)
	
	// Validate email
	if err := validation.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
		})
		return
	}
	
	// Validate password
	if err := validation.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
		})
		return
	}
	
	// Validate name
	if err := validation.ValidateName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
		})
		return
	}

	// Check for SQL injection patterns (additional security)
	if validation.CheckSQLInjection(req.Email) {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Security Error",
			Message: "Invalid input detected",
		})
		return
	}

	user := &domain.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	if err := h.userService.Register(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Registration Failed",
			Message: err.Error(),
		})
		return
	}

	// Don't return password in response
	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

// Login handles user authentication with enhanced validation
func (h *UserHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Sanitize inputs
	req.Email = validation.SanitizeInput(req.Email)
	
	// Validate email
	if err := validation.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
		})
		return
	}
	
	// Validate password is not empty
	if strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Validation Error",
			Message: "Password is required",
		})
		return
	}

	// Check for SQL injection patterns
	if validation.CheckSQLInjection(req.Email) {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Security Error",
			Message: "Invalid input detected",
		})
		return
	}

	// Get client IP and user agent
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.userService.Login(c.Request.Context(), req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Error:   "Authentication Failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	response, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Error:   "Token Refresh Failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *UserHandler) Logout(c *gin.Context) {
	// Extract session ID and token from context (set by middleware)
	sessionID := c.MustGet("session_id").(string)
	token := c.MustGet("token").(string)
	
	if sessionID == "" || token == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Session ID or token not found",
		})
		return
	}

	// Blacklist the token first
	if err := h.userService.BlacklistToken(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Logout Failed",
			Message: "Failed to blacklist token",
		})
		return
	}

	// Then logout the session
	if err := h.userService.Logout(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Logout Failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// LogoutAll handles logout from all devices
func (h *UserHandler) LogoutAll(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.userService.LogoutAll(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Logout All Failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out from all devices successfully"})
}

// GetUserSessions returns user's active sessions
func (h *UserHandler) GetUserSessions(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	sessions, err := h.userService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to retrieve user sessions",
		})
		return
	}

	c.JSON(http.StatusOK, sessions)
} 