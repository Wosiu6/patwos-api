package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimitMiddleware_WithinLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimitMiddleware(rate.Limit(10), 20))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimitMiddleware_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimitMiddleware(rate.Limit(1), 2))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			assert.Contains(t, w.Body.String(), "Rate limit exceeded")
		}
	}
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimitMiddleware(rate.Limit(1), 1))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestRateLimitMiddleware_BurstHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimitMiddleware(rate.Every(time.Second), 3))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 3 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestStrictRateLimitMiddleware_AuthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(StrictRateLimitMiddleware())
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("POST", "/login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 5 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			assert.Contains(t, w.Body.String(), "Too many attempts")
		}
	}
}

func TestIPRateLimiter_AddIP(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 20)
	
	rl := limiter.AddIP("192.168.1.1")
	assert.NotNil(t, rl)
	assert.True(t, rl.Allow())
}

func TestIPRateLimiter_GetLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 20)
	
	rl1 := limiter.GetLimiter("192.168.1.1")
	assert.NotNil(t, rl1)
	
	rl2 := limiter.GetLimiter("192.168.1.1")
	assert.NotNil(t, rl2)
	assert.Equal(t, rl1, rl2)
}

func TestIPRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(100), 200)
	
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				limiter.GetLimiter("192.168.1.1")
			}
			done <- true
		}(i)
	}
	
	for i := 0; i < 10; i++ {
		<-done
	}
}
