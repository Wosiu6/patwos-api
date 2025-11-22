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

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	gin.SetMode(cfg.GinMode)

	router := gin.New()

	router.Use(gin.Recovery())

	router.Use(gin.Logger())

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

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
