package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"products/internal/domain"
	"products/internal/service"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(userService *service.UserService, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Extract user ID and session ID
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid user ID in token",
			})
			c.Abort()
			return
		}

		sessionID, ok := claims["session_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid session ID in token",
			})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid user ID format",
			})
			c.Abort()
			return
		}

		// Validate session is still active
		isValid, err := userService.ValidateSession(c.Request.Context(), sessionID)
		if err != nil || !isValid {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Session expired or invalid",
			})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		isBlacklisted, err := userService.IsTokenBlacklisted(c.Request.Context(), tokenString)
		if err != nil || isBlacklisted {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Token has been invalidated",
			})
			c.Abort()
			return
		}

		// Check if user's session has been blacklisted by logout all
		isUserBlacklisted, err := userService.IsUserSessionBlacklisted(c.Request.Context(), userID, sessionID)
		if err != nil || isUserBlacklisted {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Session has been invalidated by logout all",
			})
			c.Abort()
			return
		}

		// Set user ID, session ID, and token in context
		c.Set("user_id", userID)
		c.Set("session_id", sessionID)
		c.Set("token", tokenString)
		c.Next()
	}
}
