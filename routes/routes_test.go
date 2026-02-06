package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestSetupRoutes_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupRoutes(r, &gorm.DB{}, &config.Config{JWTSecret: "secret"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
