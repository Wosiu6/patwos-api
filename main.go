package main

import (
	"log"
	"os"

	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/database"
	"github.com/Wosiu6/patwos-api/middleware"
	"github.com/Wosiu6/patwos-api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.LoadConfig()

	log.Printf("[DATABASE] Connecting to %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to database: %v", err)
	}
	log.Printf("[DATABASE] Connected successfully")

	log.Printf("[DATABASE] Running migrations...")
	if err := database.Migrate(db); err != nil {
		log.Fatalf("[ERROR] Failed to run migrations: %v", err)
	}
	log.Printf("[DATABASE] Migrations completed")

	gin.SetMode(cfg.GinMode)

	router := gin.New()

	router.Use(gin.Recovery())

	router.Use(gin.Logger())

	router.Use(middleware.RequestLogger())

	router.Use(middleware.SecurityHeaders())

	router.Use(middleware.RateLimitMiddleware(rate.Limit(100), 200))

	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	if len(cfg.TrustedProxies) > 0 {
		if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
			log.Printf("Warning: Failed to set trusted proxies: %v", err)
		}
	}

	router.MaxMultipartMemory = cfg.MaxRequestSize

	routes.SetupRoutes(router, db, cfg)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[STARTUP] Configuration loaded:")
	log.Printf("  - Database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("  - Port: %s", port)
	log.Printf("  - Mode: %s", cfg.GinMode)
	log.Printf("  - CORS Origins: %v", cfg.AllowedOrigins)
	log.Printf("  - Rate Limit: 100 req/s, burst: 200")
	log.Printf("[STARTUP] Starting server on port %s", port)
	log.Printf("[STARTUP] API ready - Health: http://localhost:%s/health", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("[ERROR] Failed to start server:", err)
	}
}
