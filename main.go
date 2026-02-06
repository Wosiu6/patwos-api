package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	router.Use(middleware.BodySizeLimiter(cfg.MaxRequestSize))

	router.Use(middleware.RateLimitMiddleware(rate.Limit(100), 200))

	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	if len(cfg.TrustedProxies) > 0 {
		if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
			log.Printf("Warning: Failed to set trusted proxies: %v", err)
		}
	}

	router.MaxMultipartMemory = cfg.MaxRequestSize

	routes.SetupRoutes(router, db, cfg)

	port := cfg.APIPort

	log.Printf("[STARTUP] Configuration loaded:")
	log.Printf("  - Database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("  - Port: %s", port)
	log.Printf("  - Mode: %s", cfg.GinMode)
	log.Printf("  - CORS Origins: %v", cfg.AllowedOrigins)
	log.Printf("  - Rate Limit: 100 req/s, burst: 200")
	if cfg.GinMode == "release" && cfg.DBSSLMode == "disable" {
		log.Printf("[WARNING] DB_SSLMODE is disable in release mode; enable TLS for production.")
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Printf("[STARTUP] Starting server on port %s", port)
	log.Printf("[STARTUP] API ready - Health: http://localhost:%s/health", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[SHUTDOWN] Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server shutdown failed: %v", err)
	}

	log.Printf("[SHUTDOWN] Server exited")
}
