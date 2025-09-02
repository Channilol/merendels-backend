package routes

import (
	"merendels-backend/handlers"
	"merendels-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupApprovalRoutes configura le rotte per le approvazioni con protezioni JWT
func SetupApprovalRoutes(router *gin.RouterGroup) {
	handler := handlers.NewApprovalHandler()

	// Rotte per approvals - TUTTE PROTETTE DA JWT
	approvals := router.Group("/approvals")
	approvals.Use(middleware.AuthMiddleware()) // Tutti gli endpoint richiedono autenticazione
	{
		// OPERAZIONI DI LETTURA - Accessibili a tutti gli utenti autenticati
		// Gli utenti possono vedere lo stato delle approvazioni delle proprie richieste
		
		approvals.GET("/:id", handler.GetApprovalByID)              // GET /api/approvals/:id - Singola approvazione per ID
		approvals.GET("/me", handler.GetMyApprovals)                // GET /api/approvals/me - Le mie approvazioni (come approver)
		
		// OPERAZIONI DI APPROVAZIONE - Solo hierarchy_level <= 1 (Responsabile/Capo)
		// Solo manager e responsabili possono creare, modificare ed eliminare approvazioni
		
		approvals.POST("", 
			middleware.RequireHierarchyLevel(1), 
			handler.CreateApproval)                                 // POST /api/approvals - Crea nuova approvazione
			
		approvals.PUT("/:id/status", 
			middleware.RequireHierarchyLevel(1), 
			handler.UpdateApprovalStatus)                           // PUT /api/approvals/:id/status - Aggiorna status approvazione
			
		approvals.POST("/:id/revoke", 
			middleware.RequireHierarchyLevel(1), 
			handler.RevokeApproval)                                 // POST /api/approvals/:id/revoke - Revoca approvazione
			
		approvals.DELETE("/:id", 
			middleware.RequireHierarchyLevel(1), 
			handler.DeleteApproval)                                 // DELETE /api/approvals/:id - Elimina approvazione
		
		// OPERAZIONI AMMINISTRATIVE AVANZATE - Solo hierarchy_level <= 1
		
		approvals.GET("", 
			middleware.RequireHierarchyLevel(1), 
			handler.GetAllApprovals)                                // GET /api/approvals - Tutte le approvazioni (admin/manager)
			
		approvals.GET("/status/:status", 
			middleware.RequireHierarchyLevel(1), 
			handler.GetApprovalsByStatus)                           // GET /api/approvals/status/:status - Approvazioni per status
			
		approvals.GET("/statistics", 
			middleware.RequireHierarchyLevel(1), 
			handler.GetApprovalStatistics)                          // GET /api/approvals/statistics - Statistiche approvazioni
	}

	// Rotte specifiche per richieste con approvazioni - integrate con le requests
	// Queste rotte sono accessibili a tutti gli utenti autenticati per consultazione
	requests := router.Group("/requests")
	requests.Use(middleware.AuthMiddleware())
	{
		requests.GET("/:request_id/approvals", 
			handler.GetApprovalsByRequestID)                        // GET /api/requests/:request_id/approvals - Approvazioni per richiesta
			
		requests.GET("/:request_id/approval-status", 
			handler.GetRequestApprovalStatus)                       // GET /api/requests/:request_id/approval-status - Status approvazione richiesta
	}
}