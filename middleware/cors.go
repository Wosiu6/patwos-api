package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin == "" {
			c.Next()
			return
		}

		allowed := false
		allowedOrigin := ""
		if len(allowedOrigins) > 0 {
			for _, ao := range allowedOrigins {
				if ao == "*" {
					allowed = true
					allowedOrigin = "*"
					break
				} else if ao == origin {
					allowed = true
					allowedOrigin = origin
					break
				}
			}
		}

		if !allowed {
			gin.DefaultWriter.Write([]byte("[CORS-BLOCKED] Origin: " + origin + " | Path: " + c.Request.URL.Path + " | AllowedOrigins: " + fmt.Sprint(allowedOrigins) + " | Status: 403\n"))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Origin not allowed",
			})
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		if allowedOrigin != "*" {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
