package handlers

import (
	"merendels-backend/middleware"
	"merendels-backend/models"
	"merendels-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler crea una nuova istanza dell'handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(),
	}
}

// Login gestire POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var request models.LoginRequest

	// Binding del JSON alla reques
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Chiama il service
	response, err := h.authService.Login(&request)
	if err != nil {
		// Gestione errori business
		switch err.Error() {
		case "email and password are required":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email and password are required",
			})
		case "invalid email or password":
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
		case "too many failed attempts, please try again later":
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many failed attempts, please try again later",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	// Login riuscito
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data": response,
	})
}

// Register gestisce POST api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	// Struct per ricevere i dati completi di registrazione
	var requestPayload struct {
		Name string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		RoleID *int `json:"role_id" binding:"required"`
		ManagerID *int `json:"manager_id" binding:"required"`
	}

	// Binding del JSON
	if err := c.ShouldBindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
	}

	// Trasforma in strutture separate per il service
	authRequest := &models.CreateAuthCredentialRequest{
		Password: requestPayload.Password,
	}

	userRequest := &models.CreateUserRequest{
		Name: requestPayload.Name,
		Email: requestPayload.Email,
		RoleID: requestPayload.RoleID,
		ManagerID: requestPayload.ManagerID,
	}

	// Chiama il service
	response, err := h.authService.Register(authRequest, userRequest)
	if err != nil {
		// Gestione errori business
		switch err.Error() {
			case "password cannot be empty":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password cannot be empty",
			})
		case "password must be at least 6 characters":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password must be at least 6 characters",
			})
		case "name and email are required":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Name and email are required",
			})
		case "email already registered":
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already registered",
			})
		case "invalid role_id":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid role ID",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	// Registrazione riuscita
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"data": response,
	})
}

// ChangePassword gestisce PUT /api/auth/change-password (protetto da JWT)
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Estrae user_id dal JWT Token (middleware lo mette nel context)
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Struct per ricevere la richiesta
	var request struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	// Binding del JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Chiama il service
	err := h.authService.ChangePassword(userID, request.CurrentPassword, request.NewPassword)
	if err != nil {
		// Gestione errori business
		switch err.Error() {
			case "password must not be empty":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password must not be empty",
			})
		case "password must have at least 6 characters":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password must have at least 6 characters",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
		}
		return
	}

	// Cambio password riuscito
	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}