package router

import (
	"products/internal/service"
	"products/cmd/api/internal/handler"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures the application routes
func SetupRouter(userService *service.UserService, productService *service.ProductService, jwtSecret string) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"message": "Products CRUD API is running",
		})
	})

	// Create handlers
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		public.POST("/auth/register", userHandler.Register)
		public.POST("/auth/login", userHandler.Login)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(handler.AuthMiddleware(userService, jwtSecret))
	{
		// Authentication routes
		auth := protected.Group("/auth")
		{
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/logout", userHandler.Logout)
			auth.POST("/logout-all", userHandler.LogoutAll)
			auth.GET("/sessions", userHandler.GetUserSessions)
		}

		// Product routes
		products := protected.Group("/products")
		{
			products.POST("/", productHandler.Create)
			products.GET("/", productHandler.GetAllByUser)
			products.GET("/filtered", productHandler.GetProductsWithFilters)
			products.GET("/cursor", productHandler.GetProductsWithCursor)
			products.GET("/stats", productHandler.GetProductStats)
			products.GET("/:id", productHandler.GetByID)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
		}
	}

	return router
} 