package middleware

import (
	"merendels-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifica il JWT Token in ogni richiesta
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Estrae header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			// Ferma la richiesta
			c.Abort()
			return
		}

		tokenString, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Valida il token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Salvo i claims nel context di Gin cosí da poterci accedere con gli handler
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role_id", claims.RoleID)
		c.Set("hierarchy_level", claims.HierarchyLevel)
		c.Set("claims", claims)

		// Autorizzazione ok, quindi continua
		c.Next()
	}
}

// RequireHierarchyLevel middleware per autorizzazioni basate su hierarchy_level
func RequireHierarchyLevel(minLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Recupera hierarchy_level dal context (dopo AuthMiddleware)
		hierarchy_level, exists := c.Get("hierarchy_level")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User hierarchy level not found",
			})
			c.Abort()
			return
		}

		// Controlla il tipo (potrebbe essere nil)
		userLevel, ok := hierarchy_level.(*int)
		if !ok || userLevel == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User has no hierarchy level assigned",
			})
			c.Abort()
			return
		}

		// Verifica il livello minimo richiesto, se é maggiore allora non va bene (es.: Dipendente livello 3, Responsabile livello 2)
		if *userLevel > minLevel {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficent hierarchy level",
			})
			c.Abort()
			return
		}

		// Hierarchy level validato
		c.Next()
	}
}

//* Funzioni helper per handlers

// GetUserIdFromContext estrae user_id dal context
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}

// GetUserEmailFromContext estrae email dal context
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}

	emailStr, ok := email.(string)
	return emailStr, ok
}

// GetUserClaimsFromContext estrae tutti i claims dal context
func GetUserClaimsFromContext(c *gin.Context) (*utils.JWTClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*utils.JWTClaims)
	return userClaims, ok
}