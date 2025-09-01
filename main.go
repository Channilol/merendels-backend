package main

import (
	"log"
	"merendels-backend/config"
	"merendels-backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connessione al database
	config.ConnectDatabase()
	defer config.DB.Close()

	// Setup Gin router
	router := gin.Default()

	// Middleware CORS per permette richieste dal frontend
	router.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods","GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Merendels Backend is running",
			"version": "0.0.1",
		})
	})

	// Setup API Routes group
	api := router.Group("/api")
	{
		routes.SetupAuthRoutes(api)		// Rotte autenticazione: /api/auth/*
		routes.SetupUserRoleRoutes(api)		//Rotte user roles: /api/user-roles/*
	}

	// Avvio server
	log.Println("Server starting on port 8080")
	err := router.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
