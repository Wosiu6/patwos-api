package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		if status >= 400 {
			gin.DefaultWriter.Write([]byte(
				"[REQUEST] " +
					method + " " +
					path +
					" | Status: " + http.StatusText(status) +
					" | IP: " + ip +
					" | Latency: " + latency.String() +
					"\n",
			))
		}
	}
}
