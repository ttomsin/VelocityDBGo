package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"VelocityDBGo/internal/database"
	"VelocityDBGo/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title           VelocityDBGo API
// @version         1.0
// @description     A multi-tenant JSONB document store for students.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer " followed by your JWT token.

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        X-API-Key
// @description                 Public API Key for a specific Project.

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}

	// Initialize database
	database.ConnectDB()

	// Setup Gin router
	r := gin.Default()

	// Enable CORS for all origins, allowing frontend apps to connect
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupRoutes(r)

	// Start self-pinging background worker (to keep Render free tier awake)
	go startSelfPing()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting VelocityDBGo on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func startSelfPing() {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		log.Println("APP_URL not set, self-pinging disabled.")
		return
	}

	// Ping every 14 minutes (Render sleeps after 15 mins of inactivity)
	ticker := time.NewTicker(14 * time.Minute)
	log.Printf("Self-pinging background worker started for: %s/health", appURL)

	for range ticker.C {
		resp, err := http.Get(appURL + "/health")
		if err != nil {
			log.Printf("Self-ping failed: %v", err)
			continue
		}
		resp.Body.Close()
		log.Printf("Self-ping successful: %s", resp.Status)
	}
}
