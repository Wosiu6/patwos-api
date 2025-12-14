package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter

	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()
	return limiter
}

func RateLimitMiddleware(rateLimit rate.Limit, burstSize int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rateLimit, burstSize)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.Writer.Header().Set("X-RateLimit-Limit", "100")
			c.Writer.Header().Set("X-RateLimit-Remaining", "0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			gin.DefaultWriter.Write([]byte("[RATE-LIMIT] IP: " + ip + " | Path: " + c.Request.URL.Path + " | Method: " + c.Request.Method + " | Status: 429\n"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func StrictRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Every(time.Minute), 5)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.Writer.Header().Set("X-RateLimit-Limit", "5")
			c.Writer.Header().Set("X-RateLimit-Remaining", "0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many attempts. Please try again later.",
			})
			gin.DefaultWriter.Write([]byte("[STRICT-RATE-LIMIT] IP: " + ip + " | Path: " + c.Request.URL.Path + " | Status: 429\n"))
			c.Abort()
			return
		}

		c.Next()
	}
}
