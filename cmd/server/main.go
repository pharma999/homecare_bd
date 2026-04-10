package main

import (
	"log"

	"home_care_backend/internal/config"
	"home_care_backend/internal/database"
	"home_care_backend/internal/middleware"
	"home_care_backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config from .env
	config.Load()

	// Set Gin mode
	gin.SetMode(config.AppConfig.GinMode)

	// Connect to MongoDB
	database.Connect()
	defer database.Disconnect()

	// Create router
	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": config.AppConfig.AppName})
	})

	// Register all API routes
	routes.Setup(r)

	addr := ":" + config.AppConfig.Port
	log.Printf("🚀 %s server starting on %s", config.AppConfig.AppName, addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
