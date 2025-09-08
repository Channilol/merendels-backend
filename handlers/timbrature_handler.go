package handlers

import (
	"merendels-backend/config"
	"merendels-backend/middleware"
	"merendels-backend/models"
	"merendels-backend/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TimbratureHandler struct {
	service *services.TimbratureService
}

// NewTimbratureHandler crea una nuova istanza dell'handler
func NewTimbratureHandler() *TimbratureHandler {
	return &TimbratureHandler{
		service: services.NewTimbratureService(),
	}
}

// CreateTimbrature gestisce POST /api/timbrature
func (h *TimbratureHandler) CreateTimbrature (c *gin.Context) {
	// Estrae user_id dal JWT Token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var request models.CreateTimbratureRequest

	// Binding del JSON della request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Chiama il servizio (utilizzando lo userID del token e non della request)
	response, err := h.service.CreateTimbrature(userID, &request)
	if err != nil {
		// Gestione errore specifici
		switch err.Error() {
		case "invalid action type":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid action type. Use ENTRATA or USCITA",
			})
		case "invalid location":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid location. Use UFFICIO or SMART",
			})
		case "cannot enter twice in a row - you must exit first":
			c.JSON(http.StatusConflict, gin.H{
				"error": "You already entered today. You must exit first",
			})
		case "cannot exit twice in a row - you must enter first":
			c.JSON(http.StatusConflict, gin.H{
				"error": "You already exited. You must enter first",
			})
		case "first timbratura must be ENTRATA":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Your first timbratura must be ENTRATA",
			})
		default:
			// Controllo per errori che contengono pattern specifici
			if contains := err.Error(); len(contains) > 20 && contains[:20] == "you already have a " {
				c.JSON(http.StatusConflict, gin.H{
					"error": err.Error(),
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"details": err.Error(),
				})
			}
		}
		return
	}

	email, _ := middleware.GetUserEmailFromContext(c)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Timbrature created successfully",
		"data": response,
		"created_by": email,
	})
}

// GetMyTimbrature gestisce GET /api/timbrature/me
func (h *TimbratureHandler) GetMyTimbrature(c *gin.Context) {
	// Estrae user_id dal JWT Token
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

	// Chiama il servizio
	responses, err := h.service.GetUserTimbrature(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch your timbrature",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Timbrature fetched successfully",
		"data": responses,
		"count": len(responses),
		"pagination": gin.H{
			"limit": limit,
			"offset": offset,
		},
	})
}

// GetMyTodayTimbrature gestisce GET /api/timbrature/me/today
func (h *TimbratureHandler) GetMyTodayTimbrature(c *gin.Context) {
	// Estrae user_id dal JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Chiama il service
	responses, err := h.service.GetTodayTimbrature(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch today's timbrature",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Today's timbrature fetched successfully",
		"data": responses,
		"date": time.Now().Format("2006-01-02"),
		"count": len(responses),
	})
}

// GetMyTimbratureByDate gestisce GET /api/timbrature/me/date/:date
func (h *TimbratureHandler) GetMyTimbratureByDate(c *gin.Context) {
	// Estrae user_id dal JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Estrae data dal parametro URL
	dateParam := c.Param("date")
	date, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Chiama il service
	responses, err := h.service.GetUserTimbratureByDate(userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch timbrature for date",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Timbrature for date fetched successfully",
		"data": responses,
		"date": dateParam,
		"count": len(responses),
	})
}

// GetMyWorkingStatus gestisce GET /api/timbrature/me/status
func (h *TimbratureHandler) GetMyWorkingStatus(c *gin.Context) {
	// Estrae user_id dal JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Chiama il service
	status, err := h.service.GetWorkingStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch working status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Working status fetched successfully",
		"data": status,
	})
}

// GetMyLastTimbratura gestisce GET /api/timbrature/me/last
func (h *TimbratureHandler) GetMyLastTimbrature(c *gin.Context) {
	// Estrae user_id dal JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Chiama il service
	lastTimbratura, err := h.service.GetLastTimbrature(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch last timbratura",
			"details": err.Error(),
		})
		return
	}

	if lastTimbratura == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "No previous timbrature found",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Last timbratura fetched successfully",
		"data": lastTimbratura,
	})
}

// GetAllTimbrature gestisce GET /api/timbrature (solo per admin/manager)
func (h *TimbratureHandler) GetAllTimbrature(c *gin.Context) {
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
	responses, err := h.service.GetAllTimbrature(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch all timbrature",
			"details": err.Error(),
		})
		return
	}

	// Log per audit
	adminEmail, _ := middleware.GetUserEmailFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "All timbrature fetched successfully",
		"data": responses,
		"count": len(responses),
		"pagination": gin.H{
			"limit": limit,
			"offset": offset,
		},
		"accessed_by": adminEmail,
	})
}

func (h *TimbratureHandler) GetEmployeesStatus(c *gin.Context) {
	//TODO: da spostare query nella repository
	query := `SELECT id, name, email, role_id, manager_id FROM users ORDER BY name`
	rows, err := config.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var result []gin.H
	today := time.Now().Format("2006-01-02")

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.RoleID, &user.ManagerID)
		if err != nil {
			continue
		}

		// Prendo l'ultima timbratura dell'utente oggi
		lastTimbrature, _ := h.service.GetLastTimbrature(user.ID)
		
		isWorking := false
		workMode := "unknown"
		
		if lastTimbrature != nil {
			// Controllo se è di oggi
			lastDate := lastTimbrature.Timestamp.Format("2006-01-02")
			if lastDate == today {
				isWorking = lastTimbrature.ActionType == models.ActionEnter
				workMode = string(lastTimbrature.Location) // "UFFICIO" o "SMART"
			}
		}

		result = append(result, gin.H{
			"id": user.ID,
			"name": user.Name,
			"is_working": isWorking,
			"work_mode": workMode,
		})
	}

	c.JSON(http.StatusOK, result)
}

// DeleteTimbratura gestisce DELETE /api/timbrature/:id (solo per admin)
func (h *TimbratureHandler) DeleteTimbratura(c *gin.Context) {
	// Estrae ID dal parametro URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	// Log per audit
	adminEmail, _ := middleware.GetUserEmailFromContext(c)
	adminID, _ := middleware.GetUserIDFromContext(c)

	// Chiama il service
	err = h.service.DeleteTimbratura(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete timbratura",
			"details": err.Error(),
		})
		return
	}

	// Successo
	c.JSON(http.StatusOK, gin.H{
		"message": "Timbratura deleted successfully",
		"deleted_by": gin.H{
			"user_id": adminID,
			"email": adminEmail,
		},
	})
}