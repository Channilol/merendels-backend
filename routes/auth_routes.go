package routes

import (
	"merendels-backend/handlers"
	"merendels-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.RouterGroup) {
	handler := handlers.NewAuthHandler()

	// Rotte per autenticazione
	auth := router.Group("/auth") 
	{
		// Rotte pubbliche (no middleware)
		auth.POST("/login", handler.Login) // POST /api/auth/login
		auth.POST("/register", handler.Register) // Post /api/auth/register

		// Rotte protette (middleware JWT)
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware()) // Applica middleware a tutte le rotte sotto
		{
			protected.PUT("/change-password", handler.ChangePassword)	// PUT /api/auth/change-passowrd
			protected.GET("/profile", handler.GetProfile)	// GET /api/auth/profile
			protected.POST("/logout", handler.Logout)	// POST /api/auth/logout
			protected.POST("/validate", handler.ValidateToken)	// POST /api/auth/validate
		}
	}
}