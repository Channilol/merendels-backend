package handlers

import (
	"merendels-backend/middleware"
	"merendels-backend/models"
	"merendels-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApprovalHandler struct {
	approvalService *services.ApprovalService
}

// NewApprovalHandler crea una nuova istanza dell'handler
func NewApprovalHandler() *ApprovalHandler {
	return &ApprovalHandler{
		approvalService: services.NewApprovalService(),
	}
}

// CreateApproval gestisce POST /api/approvals
func (h *ApprovalHandler) CreateApproval(c *gin.Context) {
	// Estrae user_id dal JWT Token (questo sarà l'approver)
	approverID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var request models.CreateApprovalRequest

	// Binding del JSON della request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Override dell'ApproverID con quello dal token JWT (sicurezza)
	request.ApproverID = approverID

	// Chiama il service
	createdApproval, err := h.approvalService.CreateApproval(approverID, &request)
	if err != nil {
		// Gestione errori specifici
		switch err.Error() {
		case "ID richiesta non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request ID",
			})
		case "status approvazione non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid approval status. Use APPROVED, REJECTED, or REVOKED",
			})
		case "richiesta non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Request not found",
			})
		case "non è possibile approvare le proprie richieste":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Cannot approve your own requests",
			})
		case "hai già dato un'approvazione per questa richiesta":
			c.JSON(http.StatusConflict, gin.H{
				"error": "You have already provided an approval for this request",
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
		"message": "Approval created successfully",
		"data": createdApproval,
		"approved_by": email,
	})
}

// GetAllApprovals gestisce GET /api/approvals (solo per admin/manager)
func (h *ApprovalHandler) GetAllApprovals(c *gin.Context) {
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
	approvals, err := h.approvalService.GetAllApprovals(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch approvals",
			"details": err.Error(),
		})
		return
	}

	// Log per audit
	adminEmail, _ := middleware.GetUserEmailFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Approvals fetched successfully",
		"data": approvals,
		"count": len(approvals),
		"pagination": gin.H{
			"limit": limit,
			"offset": offset,
		},
		"accessed_by": adminEmail,
	})
}

// GetApprovalByID gestisce GET /api/approvals/:id
func (h *ApprovalHandler) GetApprovalByID(c *gin.Context) {
	// Estrae ID dal parametro URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid approval ID format",
		})
		return
	}

	// Chiama il service
	approval, err := h.approvalService.GetApprovalByID(id)
	if err != nil {
		switch err.Error() {
		case "ID approvazione non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid approval ID",
			})
		case "approvazione non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Approval not found",
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
		"message": "Approval fetched successfully",
		"data": approval,
	})
}

// GetApprovalsByRequestID gestisce GET /api/requests/:request_id/approvals
func (h *ApprovalHandler) GetApprovalsByRequestID(c *gin.Context) {
	// Estrae request_id dal parametro URL
	requestIDParam := c.Param("request_id")
	requestID, err := strconv.Atoi(requestIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	// Chiama il service
	approvals, err := h.approvalService.GetApprovalsByRequestID(requestID)
	if err != nil {
		switch err.Error() {
		case "ID richiesta non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request ID",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch approvals for request",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Approvals for request fetched successfully",
		"data": approvals,
		"count": len(approvals),
		"request_id": requestID,
	})
}

// GetMyApprovals gestisce GET /api/approvals/me
func (h *ApprovalHandler) GetMyApprovals(c *gin.Context) {
	// Estrae user_id dal JWT token
	approverID, exists := middleware.GetUserIDFromContext(c)
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

	// Chiama il service
	approvals, err := h.approvalService.GetApprovalsByApproverID(approverID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch your approvals",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Your approvals fetched successfully",
		"data": approvals,
		"count": len(approvals),
		"pagination": gin.H{
			"limit": limit,
			"offset": offset,
		},
	})
}

// GetApprovalsByStatus gestisce GET /api/approvals/status/:status
func (h *ApprovalHandler) GetApprovalsByStatus(c *gin.Context) {
	// Estrae status dal parametro URL
	statusParam := c.Param("status")
	var status models.ApprovalStatus

	switch statusParam {
	case "APPROVED":
		status = models.ApprovalAccepted
	case "REJECTED":
		status = models.ApprovalRejected
	case "REVOKED":
		status = models.ApprovalRevoked
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Use APPROVED, REJECTED, or REVOKED",
		})
		return
	}

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
	approvals, err := h.approvalService.GetApprovalsByStatus(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch approvals by status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Approvals fetched successfully",
		"data": approvals,
		"count": len(approvals),
		"status": statusParam,
		"pagination": gin.H{
			"limit": limit,
			"offset": offset,
		},
	})
}

// UpdateApprovalStatus gestisce PUT /api/approvals/:id/status
func (h *ApprovalHandler) UpdateApprovalStatus(c *gin.Context) {
	// Estrae user_id dal JWT token
	approverID, exists := middleware.GetUserIDFromContext(c)
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
			"error": "Invalid approval ID format",
		})
		return
	}

	var updateRequest struct {
		Status   string  `json:"status" binding:"required"`
		Comments *string `json:"comments"`
	}

	// Binding del JSON della request
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Converte lo status string in ApprovalStatus
	var status models.ApprovalStatus
	switch updateRequest.Status {
	case "APPROVED":
		status = models.ApprovalAccepted
	case "REJECTED":
		status = models.ApprovalRejected
	case "REVOKED":
		status = models.ApprovalRevoked
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Use APPROVED, REJECTED, or REVOKED",
		})
		return
	}

	// Chiama il service
	updatedApproval, err := h.approvalService.UpdateApprovalStatus(id, approverID, status, updateRequest.Comments)
	if err != nil {
		// Gestione errori specifici
		switch err.Error() {
		case "ID approvazione non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid approval ID",
			})
		case "approvazione non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Approval not found",
			})
		case "non autorizzato a modificare questa approvazione":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Not authorized to modify this approval",
			})
		case "non è possibile modificare un'approvazione già accettata (solo revoca)":
			c.JSON(http.StatusConflict, gin.H{
				"error": "Cannot modify an approved approval (only revocation allowed)",
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
		"message": "Approval status updated successfully",
		"data": updatedApproval,
		"updated_by": email,
	})
}

// RevokeApproval gestisce POST /api/approvals/:id/revoke
func (h *ApprovalHandler) RevokeApproval(c *gin.Context) {
	// Estrae user_id dal JWT token
	approverID, exists := middleware.GetUserIDFromContext(c)
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
			"error": "Invalid approval ID format",
		})
		return
	}

	var revokeRequest struct {
		Reason string `json:"reason"`
	}

	// Binding del JSON della request (reason è opzionale)
	c.ShouldBindJSON(&revokeRequest)

	// Chiama il service
	revokedApproval, err := h.approvalService.RevokeApproval(id, approverID, revokeRequest.Reason)
	if err != nil {
		// Gestione errori specifici
		switch err.Error() {
		case "ID approvazione non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid approval ID",
			})
		case "approvazione non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Approval not found",
			})
		case "non autorizzato a revocare questa approvazione":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Not authorized to revoke this approval",
			})
		case "è possibile revocare solo approvazioni accettate":
			c.JSON(http.StatusConflict, gin.H{
				"error": "Can only revoke approved approvals",
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
		"message": "Approval revoked successfully",
		"data": revokedApproval,
		"revoked_by": email,
	})
}

// DeleteApproval gestisce DELETE /api/approvals/:id
func (h *ApprovalHandler) DeleteApproval(c *gin.Context) {
	// Estrae user_id dal JWT token
	approverID, exists := middleware.GetUserIDFromContext(c)
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
			"error": "Invalid approval ID format",
		})
		return
	}

	// Chiama il service
	err = h.approvalService.DeleteApproval(id, approverID)
	if err != nil {
		// Gestione errori specifici
		switch err.Error() {
		case "ID approvazione non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid approval ID",
			})
		case "approvazione non trovata":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Approval not found",
			})
		case "non autorizzato a eliminare questa approvazione":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Not authorized to delete this approval",
			})
		case "non è possibile eliminare un'approvazione accettata (usa revoca)":
			c.JSON(http.StatusConflict, gin.H{
				"error": "Cannot delete an approved approval (use revoke instead)",
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
		"message": "Approval deleted successfully",
		"deleted_by": email,
	})
}

// GetRequestApprovalStatus gestisce GET /api/requests/:request_id/approval-status
func (h *ApprovalHandler) GetRequestApprovalStatus(c *gin.Context) {
	// Estrae request_id dal parametro URL
	requestIDParam := c.Param("request_id")
	requestID, err := strconv.Atoi(requestIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	// Chiama il service
	status, err := h.approvalService.GetRequestApprovalStatus(requestID)
	if err != nil {
		switch err.Error() {
		case "ID richiesta non valido":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request ID",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch approval status",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Request approval status fetched successfully",
		"data": status,
	})
}

// GetApprovalStatistics gestisce GET /api/approvals/statistics
func (h *ApprovalHandler) GetApprovalStatistics(c *gin.Context) {
	// Chiama il service
	stats, err := h.approvalService.GetApprovalStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch approval statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Approval statistics fetched successfully",
		"data": stats,
	})
}