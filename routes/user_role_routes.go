package routes

import (
	"merendels-backend/handlers"
	"merendels-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserRoleRoutes configura le rotte per user_roles con protezioni JWT
func SetupUserRoleRoutes(router *gin.RouterGroup) {
	handler := handlers.NewUserRoleHandler()

	// Rotte per user_roles TUTTE protette da JWT
	userRoles := router.Group("/user-roles")
	userRoles.Use(middleware.AuthMiddleware()) // Tutti gli endpoint richiedono l'autenticazione
	{
		// OPERAZIONI DI LETTURA - Accessibili a tutti gli utenti autenticati
		userRoles.GET("", handler.GetAllUserRoles)      // GET /api/user-roles
		userRoles.GET("/:id", handler.GetUserRoleByID)  // GET /api/user-roles/:id
		
		// OPERAZIONI DI SCRITTURA - Solo hierarchy_level >= 2 (Manager+)
		userRoles.POST("", 
			middleware.RequireHierarchyLevel(2), 
			handler.CreateUserRole)  // POST /api/user-roles
			
		userRoles.PUT("/:id", 
			middleware.RequireHierarchyLevel(2), 
			handler.UpdateUserRole)  // PUT /api/user-roles/:id
			
		userRoles.DELETE("/:id", 
			middleware.RequireHierarchyLevel(2), 
			handler.DeleteUserRole)  // DELETE /api/user-roles/:id

	}
}