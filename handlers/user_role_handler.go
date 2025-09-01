package handlers

import (
	"merendels-backend/middleware"
	"merendels-backend/models"
	"merendels-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserRoleHanlder struct {
	service *services.UserRoleService
}

// Nuova istanza dell'handler
func NewUserRoleHandler() *UserRoleHanlder {
	return &UserRoleHanlder{
		service: services.NewUserRoleRepository(),
	}
}

func (h *UserRoleHanlder) CreateUserRole(ctx *gin.Context) {
	var request models.CreateUserRoleRequest

	// Binding JSON request
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Log dell'utente che sta creando il ruoto
	userID, _ := middleware.GetUserIDFromContext(ctx)
	email, _ := middleware.GetUserEmailFromContext(ctx)

	// Chiama il service
	response, err := h.service.CreateUserRole(&request)
	if err != nil {
		// Determina il tipo di errore
		switch err.Error() {
			case "name cannot be empty":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Name cannot be empty",
			})
		case "hierarchy level must be positive":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Hierarchy level must be positive",
			})
		case "hierarchy level already exists":
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Hierarchy level already exists",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	// Log successo
	ctx.Header("X-Created-By", email)

	// Successo
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User role created successfully",
		"data": response,
		"created_by": map[string]interface{}{
			"user_id": userID,
			"email": email,
		},
	})
}

// GetAllUserRoles gestisce GET /api/user-roles
func (h *UserRoleHanlder) GetAllUserRoles(ctx *gin.Context) {
	responses, err := h.service.GetAllUserRoles()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user roles",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User roles fetched successfully",
		"data": responses,
		"count": len(responses),
	})
}

// GetUserRoleByID gestisce GET /api/user-roles/:id
func (h *UserRoleHanlder) GetUserRoleByID(ctx *gin.Context) {
	// Estrae ID dal parametro URL
	idParam	:= ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
	return
	}

	response, err := h.service.GetUserRoleByID(id)
	if err != nil {
		switch err.Error() {
		case "invalid ID: must be greater than 0":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID: must be greater than 0",
			})
		case "user role not found":
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "User role not found",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User role fetched successfully",
		"data": response,
	})
}

// UpdateUserRole gestisce PUT /api/user-roles/:id
func (h *UserRoleHanlder) UpdateUserRole(ctx *gin.Context) {
	// Estrae ID dal parametro URL
	idParam	:= ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
	return
	}

	var request models.CreateUserRoleRequest

	
	// Binding JSON request
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Log dell'utente che sta aggiornando il ruolo
	userID, _ := middleware.GetUserIDFromContext(ctx)
	email, _ := middleware.GetUserEmailFromContext(ctx)

	// Chiama il service
	response, err := h.service.UpdateUserRole(id, &request)
	if err != nil {
		// Gestione errori business
		switch err.Error() {
		case "invalid ID: must be greater than 0":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID: must be greater than 0",
			})
		case "name cannot be empty":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Name cannot be empty",
			})
		case "hierarchy level must be positive":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Hierarchy level must be positive",
			})
		case "user role not found":
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "User role not found",
			})
		case "hierarchy level already exists":
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Hierarchy level already exists",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	// Successo
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"data": response,
		"updated_by": map[string]interface{}{
			"user_id": userID,
			"email": email,
		},
	})
}

func (h *UserRoleHanlder) DeleteUserRole(ctx *gin.Context) {
	// Estrae ID dal parametro URL
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	// Log dell'utente che sta eliminando il ruolo
	userID, _ := middleware.GetUserIDFromContext(ctx)
	email, _ := middleware.GetUserEmailFromContext(ctx)

	// Chiama il service
	_, err = h.service.DeleteUserRole(id)
	if err != nil {
		// Gestione errori business
		switch err.Error() {
		case "invalid ID: must be greater than 0":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID: must be greater than 0",
			})
		case "user role not found":
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "User role not found",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	// Successo
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User role deleted successfully",
		"deleted_by": map[string]interface{}{
			"user_id": userID,
			"email": email,
		},
	})
}