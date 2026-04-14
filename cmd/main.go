package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/dansonserge/DaFileService/config"
	"github.com/dansonserge/DaFileService/controllers"
	"github.com/dansonserge/DaFileService/middlewares"
	"github.com/dansonserge/DaFileService/routes"
	"github.com/dansonserge/DaFileService/services"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize Institutional MinIO Object Storage
	minioSvc, err := services.NewMinioService(cfg)
	if err != nil {
		log.Fatalf("Critical Hub Initialization Failure: %v", err)
	}

	// Initialize Services
	pdfSvc := services.NewPDFService(minioSvc, cfg.AuthServiceURL)

	// Start Background NATS Listener for Marketplace Events
	listener, err := services.NewNATSListener(pdfSvc, minioSvc, cfg.DefaultBucket)
	if err != nil {
		log.Printf("⚠️ Network Sync Warning: NATS connectivity deferred: %v", err)
	} else {
		log.Println("✅ Document Node listening for Marketplace settlements...")
		go listener.Listen()
	}

	// Initialize Controllers
	fileCtrl := controllers.NewFileController(minioSvc, cfg.DefaultBucket)

	// Setup Server
	router := gin.Default()

	// Global Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.CORSMiddleware())

	// Register Organizational Routes
	routes.SetupRoutes(router, fileCtrl)

	log.Printf("DaFileService Secure Node starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to establish server listener: %v", err)
	}
}
