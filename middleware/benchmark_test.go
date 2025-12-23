package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func BenchmarkAuthMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(&testing.T{})
	cfg := testhelpers.GetTestConfig()
	user := testhelpers.CreateTestUser(&testing.T{}, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	token := testhelpers.GenerateTestJWT(user.ID, user.State, user.Role, cfg.JWTSecret)

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRateLimitMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(rate.Limit(1000), 2000))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkSecurityHeaders(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCORSMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware([]string{"http://localhost:3000"}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
