package handlers

import (
	"merendels-backend/middleware"
	"merendels-backend/models"
	"merendels-backend/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type RequestHandler struct {
	requestService *services.RequestService
}

// NewRequestHandler crea una nuova istanza dell'handler
func NewRequestHandler() *RequestHandler {
	return &RequestHandler{
		requestService: services.NewRequestService(),
	}
}

// CreateRequest gestisce POST /api/requests
func (h *RequestHandler) CreateRequest(c *gin.Context) {
	// Estrae user_id dal JWT Token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var request models.CreateRequest

	// Binding del JSON della request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Override dell'UserID con quello dal token JWT (sicurezza)
	request.UserID = userID

	// Chiama il service
	createdRequest, err := h.requestService.CreateRequest(userID, &request)
	if err != nil {
		// Gestione errori specifici
		switch err.Error() {
		case "data inizio non può essere successiva alla data fine":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Start date cannot be after end date",
			})
		case "non è possibile richiedere ferie per date passate":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot request leave for past dates",
			})
		case "tipo richiesta non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request type. Use FERIE or PERMESSO",
			})
		case "la richiesta deve coprire almeno un giorno lavorativo":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Request must cover at least one working day",
			})
		case "non è possibile richiedere più di 30 giorni consecutivi di ferie":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot request more than 30 consecutive days of holidays",
			})
		case "non è possibile richiedere più di 5 giorni consecutivi di permessi":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot request more than 5 consecutive days of permits",
			})
		case "esiste già una richiesta per questo periodo":
			c.JSON(http.StatusConflict, gin.H{
				"error": "A request already exists for this period",
			})
		case "saldo ferie insufficiente per questa richiesta":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient leave balance for this request",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	email, _ := middleware.GetUserEmailFromContext(c)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Request created successfully",
		"data": createdRequest,
		"created_by": email,
	})
}

// GetAllRequests gestisce GET /api/requests (solo per admin/manager)
func (h *RequestHandler) GetAllRequests(c *gin.Context) {
	// Parametri di paginazione
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Chiama il service
	requests, err := h.requestService.GetAllRequests(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch requests",
			"details": err.Error(),
		})
		return
	}

	// Log per audit
	adminEmail, _ := middleware.GetUserEmailFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Requests fetched successfully",
		"data": requests,
		"count": len(requests),
		"pagination": gin.H{
			"limit": limit,
			"offset": offset,
		},
		"accessed_by": adminEmail,
	})
}

// GetRequestByID gestisce GET /api/requests/:id
func (h *RequestHandler) GetRequestByID(c *gin.Context) {
	// Estrae ID dal parametro URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	// Chiama il service
	request, err := h.requestService.GetRequestByID(id)
	if err != nil {
		switch err.Error() {
		case "ID richiesta non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request ID",
			})
		case "richiesta non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Request not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Request fetched successfully",
		"data": request,
	})
}

// GetMyRequests gestisce GET /api/requests/me (VERSIONE AGGIORNATA)
func (h *RequestHandler) GetMyRequests(c *gin.Context) {
	// Estrae user_id dal JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parametri di paginazione opzionali
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Chiama il service AGGIORNATO
	requests, err := h.requestService.GetUserRequestsWithStatus(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch your requests",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Your requests fetched successfully",
		"data": requests,
		"count": len(requests),
		"pagination": gin.H{
			"limit": limit,
			"offset": offset,
		},
	})
}

// GetRequestsByDateRange gestisce GET /api/requests/date-range?start_date=...&end_date=...
func (h *RequestHandler) GetRequestsByDateRange(c *gin.Context) {
	// Parametri date dalla query string
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date and end_date parameters are required (YYYY-MM-DD format)",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Chiama il service
	requests, err := h.requestService.GetRequestsByDateRange(startDate, endDate)
	if err != nil {
		switch err.Error() {
		case "data inizio non può essere successiva alla data fine":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Start date cannot be after end date",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch requests by date range",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Requests fetched successfully",
		"data": requests,
		"count": len(requests),
		"date_range": gin.H{
			"start_date": startDateStr,
			"end_date": endDateStr,
		},
	})
}

// GetPendingRequests gestisce GET /api/requests/pending
func (h *RequestHandler) GetPendingRequests(c *gin.Context) {
	// Chiama il service
	requests, err := h.requestService.GetPendingRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch pending requests",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pending requests fetched successfully",
		"data": requests,
		"count": len(requests),
	})
}

// UpdateRequest gestisce PUT /api/requests/:id
func (h *RequestHandler) UpdateRequest(c *gin.Context) {
	// Estrae user_id dal JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Estrae ID dal parametro URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	var request models.CreateRequest

	// Binding del JSON della request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Override dell'UserID con quello dal token JWT (sicurezza)
	request.UserID = userID

	// Chiama il service
	updatedRequest, err := h.requestService.UpdateRequest(id, userID, &request)
	if err != nil {
		// Gestione errori specifici
		switch err.Error() {
		case "ID richiesta non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request ID",
			})
		case "richiesta non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Request not found",
			})
		case "non autorizzato a modificare questa richiesta":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Not authorized to modify this request",
			})
		case "non è possibile modificare una richiesta già approvata o rifiutata":
			c.JSON(http.StatusConflict, gin.H{
				"error": "Cannot modify a request that has been approved or rejected",
			})
		case "le nuove date si sovrappongono con un'altra richiesta esistente":
			c.JSON(http.StatusConflict, gin.H{
				"error": "New dates overlap with another existing request",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	email, _ := middleware.GetUserEmailFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Request updated successfully",
		"data": updatedRequest,
		"updated_by": email,
	})
}

// DeleteRequest gestisce DELETE /api/requests/:id
func (h *RequestHandler) DeleteRequest(c *gin.Context) {
	// Estrae user_id dal JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Estrae ID dal parametro URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	// Chiama il service
	err = h.requestService.DeleteRequest(id, userID)
	if err != nil {
		// Gestione errori specifici
		switch err.Error() {
		case "ID richiesta non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request ID",
			})
		case "richiesta non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Request not found",
			})
		case "non autorizzato a eliminare questa richiesta":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Not authorized to delete this request",
			})
		case "non è possibile eliminare una richiesta già approvata":
			c.JSON(http.StatusConflict, gin.H{
				"error": "Cannot delete an approved request",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	email, _ := middleware.GetUserEmailFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Request deleted successfully",
		"deleted_by": email,
	})
}

// GetRequestWithApprovals gestisce GET /api/requests/:id/approvals
func (h *RequestHandler) GetRequestWithApprovals(c *gin.Context) {
	// Estrae ID dal parametro URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	// Chiama il service
	requestWithApprovals, err := h.requestService.GetRequestWithApprovals(id)
	if err != nil {
		switch err.Error() {
		case "ID richiesta non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request ID",
			})
		case "richiesta non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Request not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Request with approvals fetched successfully",
		"data": requestWithApprovals,
	})
}