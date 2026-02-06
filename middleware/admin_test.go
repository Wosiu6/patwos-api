package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/gin-gonic/gin"
)

func TestAdminMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("user_role", models.UserRoleAdmin)
		AdminMiddleware(nil)(c)
		if !c.IsAborted() {
			c.Status(http.StatusOK)
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
