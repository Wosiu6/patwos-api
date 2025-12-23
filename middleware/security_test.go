package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	testhelpers.AssertSecurityHeaders(t, w)
}

func TestSecurityHeaders_AllEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	router.POST("/api/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	router.PUT("/api/articles/1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	router.DELETE("/api/articles/1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/users"},
		{"POST", "/api/login"},
		{"PUT", "/api/articles/1"},
		{"DELETE", "/api/articles/1"},
	}

	for _, endpoint := range endpoints {
		req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertSecurityHeaders(t, w)
	}
}

func TestSecurityHeaders_XFrameOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestSecurityHeaders_ContentTypeOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
}

func TestSecurityHeaders_HSTS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_CSP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
}

func TestSecurityHeaders_PermissionsPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "geolocation=(), microphone=(), camera=()", w.Header().Get("Permissions-Policy"))
}
