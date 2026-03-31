package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/database"
	"github.com/Wosiu6/patwos-api/middleware"
	"github.com/Wosiu6/patwos-api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "migrate":
		runMigrate()
	case "migrate-force":
		if len(os.Args) < 3 {
			log.Fatal("[ERROR] Usage: ./main migrate-force <version>")
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("[ERROR] Invalid version: %s", os.Args[2])
		}
		runMigrateForce(version)
	case "serve":
		runServe()
	default:
		log.Fatalf("[ERROR] Unknown command: %s (use 'migrate', 'migrate-force <version>', or 'serve')", cmd)
	}
}

func runMigrate() {
	cfg := config.LoadConfig()

	log.Printf("[DATABASE] Connecting to %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to database: %v", err)
	}

	log.Printf("[MIGRATE] Running migrations...")
	if err := database.RunMigrations(db, migrationsFS); err != nil {
		log.Fatalf("[ERROR] Migration failed: %v", err)
	}
	log.Printf("[MIGRATE] Migrations completed successfully")
}

func runMigrateForce(version int) {
	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to database: %v", err)
	}

	log.Printf("[MIGRATE] Forcing migration version to %d...", version)
	if err := database.ForceMigrationVersion(db, migrationsFS, version); err != nil {
		log.Fatalf("[ERROR] Force version failed: %v", err)
	}
	log.Printf("[MIGRATE] Forced to version %d successfully", version)
}

func runServe() {
	cfg := config.LoadConfig()

	log.Printf("[DATABASE] Connecting to %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to database: %v", err)
	}
	log.Printf("[DATABASE] Connected successfully")

	gin.SetMode(cfg.GinMode)

	router := gin.New()

	router.Use(gin.Recovery())

	router.Use(middleware.SecurityHeaders())

	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	router.Use(middleware.RateLimitMiddleware(rate.Limit(100), 200))

	router.Use(middleware.RequestTimeout(cfg.RequestTimeout))

	router.Use(middleware.BodySizeLimiter(cfg.MaxRequestSize))

	router.Use(middleware.RequestLogger())

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
