package routes

import (
	"merendels-backend/handlers"
	"merendels-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRequestRoutes configura le rotte per le richieste di ferie/permessi con protezioni JWT
func SetupRequestRoutes(router *gin.RouterGroup) {
	handler := handlers.NewRequestHandler()

	// Rotte per requests - TUTTE PROTETTE DA JWT
	requests := router.Group("/requests")
	requests.Use(middleware.AuthMiddleware()) // Tutti gli endpoint richiedono autenticazione
	{
		// IMPORTANTE: Rotte specifiche PRIMA di quelle con parametri dinamici per evitare conflitti routing
		
		// OPERAZIONI PERSONALI - Tutti gli utenti autenticati possono gestire le proprie richieste
		requests.POST("", handler.CreateRequest)                    // POST /api/requests - Crea una nuova richiesta
		requests.GET("/me", handler.GetMyRequests)                  // GET /api/requests/me - Le mie richieste
		
		// OPERAZIONI DI CONSULTA - Accessibili a tutti gli utenti autenticati
		requests.GET("/date-range", handler.GetRequestsByDateRange) // GET /api/requests/date-range?start_date=...&end_date=... - Richieste per range date
		
		// OPERAZIONI AMMINISTRATIVE - Solo hierarchy_level <= 1 (Responsabile/Capo)
		requests.GET("/pending", 
			middleware.RequireHierarchyLevel(1), 
			handler.GetPendingRequests)                             // GET /api/requests/pending - Richieste in attesa di approvazione
		
		// Gestione globale richieste (solo per manager/admin) - DEVE essere prima di /:id per evitare conflitti
		requests.GET("", 
			middleware.RequireHierarchyLevel(1), 
			handler.GetAllRequests)                                 // GET /api/requests - Tutte le richieste (admin/manager)
		
		// ROTTE CON PARAMETRI DINAMICI - Devono essere ALLA FINE per evitare conflitti
		requests.GET("/:id", handler.GetRequestByID)                // GET /api/requests/:id - Singola richiesta per ID
		requests.GET("/:id/approvals", handler.GetRequestWithApprovals) // GET /api/requests/:id/approvals - Richiesta con approvazioni
		requests.PUT("/:id", handler.UpdateRequest)                 // PUT /api/requests/:id - Aggiorna mia richiesta
		requests.DELETE("/:id", handler.DeleteRequest)              // DELETE /api/requests/:id - Elimina mia richiesta
	}
}