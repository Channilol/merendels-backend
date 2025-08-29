package main

import (
	"bytes"
	"log"
	"merendels-backend/middleware"
	"merendels-backend/utils"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🧪 Test Auth Middleware")

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Endpoint protetto con solo AuthMiddleware
	router.GET("/protected", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, _ := middleware.GetUserIDFromContext(c)
		email, _ := middleware.GetUserEmailFromContext(c)
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Access granted",
			"user_id": userID,
			"email":   email,
		})
	})

	// Endpoint protetto con hierarchy level (solo manager level 2+)
	router.POST("/admin-only", 
		middleware.AuthMiddleware(), 
		middleware.RequireHierarchyLevel(2), 
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Admin access granted",
			})
		})

	// Test 1: Richiesta senza token (dovrebbe fallire)
	log.Println("\n❌ Test 1: No Token")
	req1, _ := http.NewRequest("GET", "/protected", nil)
	resp1 := httptest.NewRecorder()
	router.ServeHTTP(resp1, req1)
	log.Printf("Status: %d", resp1.Code)
	log.Printf("Body: %s", resp1.Body.String())

	// Test 2: Token valido (dovrebbe funzionare)
	log.Println("\n✅ Test 2: Valid Token")
	
	// Genera token per utente di test
	userID := 5
	email := "mario@test.com"
	roleID := 2
	hierarchyLevel := 3
	token, err := utils.GenerateToken(userID, email, &roleID, &hierarchyLevel)
	if err != nil {
		log.Printf("❌ Errore generazione token: %v", err)
		return
	}

	req2, _ := http.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)
	log.Printf("Status: %d", resp2.Code)
	log.Printf("Body: %s", resp2.Body.String())

	// Test 3: Hierarchy level sufficiente (level 3 >= 2)
	log.Println("\n✅ Test 3: Sufficient Hierarchy Level")
	req3, _ := http.NewRequest("POST", "/admin-only", bytes.NewBuffer([]byte("{}")))
	req3.Header.Set("Authorization", "Bearer "+token)
	req3.Header.Set("Content-Type", "application/json")
	resp3 := httptest.NewRecorder()
	router.ServeHTTP(resp3, req3)
	log.Printf("Status: %d", resp3.Code)
	log.Printf("Body: %s", resp3.Body.String())

	// Test 4: Hierarchy level insufficiente
	log.Println("\n❌ Test 4: Insufficient Hierarchy Level")
	
	// Genera token per utente con level basso
	lowLevelToken, _ := utils.GenerateToken(6, "junior@test.com", &roleID, new(int)) // level 0
	
	req4, _ := http.NewRequest("POST", "/admin-only", bytes.NewBuffer([]byte("{}")))
	req4.Header.Set("Authorization", "Bearer "+lowLevelToken)
	req4.Header.Set("Content-Type", "application/json")
	resp4 := httptest.NewRecorder()
	router.ServeHTTP(resp4, req4)
	log.Printf("Status: %d", resp4.Code)
	log.Printf("Body: %s", resp4.Body.String())

	// Test 5: Token malformato
	log.Println("\n❌ Test 5: Invalid Token")
	req5, _ := http.NewRequest("GET", "/protected", nil)
	req5.Header.Set("Authorization", "Bearer token-fasullo")
	resp5 := httptest.NewRecorder()
	router.ServeHTTP(resp5, req5)
	log.Printf("Status: %d", resp5.Code)
	log.Printf("Body: %s", resp5.Body.String())

	log.Println("\n🎉 Tutti i test middleware completati!")
}