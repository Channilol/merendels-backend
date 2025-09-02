package routes

import (
	"merendels-backend/handlers"
	"merendels-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupTimbratureRoutes configura le rotte per le timbrature con protezioni JWT
func SetupTimbratureRoutes(router *gin.RouterGroup) {
	handler := handlers.NewTimbratureHandler()
	
	// Rotte per timbrature - TUTTE PROTETTE DA JWT
	timbrature := router.Group("/timbrature")
	timbrature.Use(middleware.AuthMiddleware()) // Tutti gli endpoint richiedono autenticazione
	{
		// OPERAZIONI PERSONALI - Tutti gli utenti autenticati
		// Endpoint per gestire le proprie timbrature
		timbrature.POST("", handler.CreateTimbrature) // POST /api/timbrature - Crea timbratura
		timbrature.GET("/me", handler.GetMyTimbrature) // GET /api/timbrature/me - Le mie timbrature
		timbrature.GET("/me/today", handler.GetMyTodayTimbrature) // GET /api/timbrature/me/today - Timbrature di oggi
		timbrature.GET("/me/date/:date", handler.GetMyTimbratureByDate) // GET /api/timbrature/me/date/2025-01-15
		timbrature.GET("/me/status", handler.GetMyWorkingStatus) // GET /api/timbrature/me/status - Stato lavorativo
		timbrature.GET("/me/last", handler.GetMyLastTimbrature) // GET /api/timbrature/me/last - Ultima timbratura
		
		// OPERAZIONI AMMINISTRATIVE - Solo hierarchy_level <= 1 (Responsabile/Capo)
		timbrature.GET("", 
			middleware.RequireHierarchyLevel(1), 
			handler.GetAllTimbrature)  // GET /api/timbrature - Tutte le timbrature
			
		timbrature.DELETE("/:id", 
			middleware.RequireHierarchyLevel(1), 
			handler.DeleteTimbratura)  // DELETE /api/timbrature/:id - Elimina timbratura
	}
}