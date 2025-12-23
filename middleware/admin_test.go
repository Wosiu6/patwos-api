package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdminMiddleware_NoUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)

	router := gin.New()
	router.Use(AdminMiddleware(db))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminMiddleware_RegularUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", models.UserRoleUser)
		c.Next()
	})
	router.Use(AdminMiddleware(db))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminMiddleware_AdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", models.UserRoleAdmin)
		c.Next()
	})
	router.Use(AdminMiddleware(db))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminMiddleware_InvalidRoleType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "invalid")
		c.Next()
	})
	router.Use(AdminMiddleware(db))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
