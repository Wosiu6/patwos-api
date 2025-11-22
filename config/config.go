package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBHost         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBPort         string
	JWTSecret      string
	APIPort        string
	GinMode        string
	AllowedOrigins []string
	TrustedProxies []string
	MaxRequestSize int64
}

func LoadConfig() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}

	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBUser:         getEnv("DB_USER", "patwos"),
		DBName:         getEnv("DB_NAME", "patwos_db"),
		DBPort:         getEnv("DB_PORT", "5432"),
		JWTSecret:      jwtSecret,
		DBPassword:     dbPassword,
		APIPort:        getEnv("API_PORT", "8080"),
		GinMode:        getEnv("GIN_MODE", "debug"),
		AllowedOrigins: getEnvArray("ALLOWED_ORIGINS", []string{"*"}),
		TrustedProxies: getEnvArray("TRUSTED_PROXIES", []string{}),
		MaxRequestSize: getEnvInt64("MAX_REQUEST_SIZE", 10485760),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvArray(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		var result []string
		for _, v := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(v)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
