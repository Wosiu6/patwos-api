package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	normalizedAllowed := make(map[string]struct{}, len(allowedOrigins))
	allowAll := false
	for _, ao := range allowedOrigins {
		normalized := normalizeOrigin(ao)
		if normalized == "" {
			continue
		}
		if normalized == "*" {
			allowAll = true
			break
		}
		normalizedAllowed[normalized] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin == "" {
			c.Next()
			return
		}

		normalizedOrigin := normalizeOrigin(origin)

		allowed := false
		allowedOrigin := ""
		if allowAll {
			allowed = true
			allowedOrigin = "*"
		} else if normalizedOrigin != "" {
			if _, ok := normalizedAllowed[normalizedOrigin]; ok {
				allowed = true
				allowedOrigin = normalizedOrigin
			}
		}

		if !allowed {
			log.Printf("[CORS-BLOCKED] Origin: %s | Path: %s | AllowedOrigins: %s | Status: 403", origin, c.Request.URL.Path, fmt.Sprint(allowedOrigins))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Origin not allowed",
			})
			return
		}

		setVaryHeader(c, "Origin")
		if c.Request.Method == http.MethodOptions {
			setVaryHeader(c, "Access-Control-Request-Method")
			setVaryHeader(c, "Access-Control-Request-Headers")
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		if allowedOrigin != "*" {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
		if requestHeaders == "" {
			requestHeaders = "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin, Cache-Control, X-Requested-With"
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", requestHeaders)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(strings.TrimRight(origin, "/"))
	origin = strings.Trim(origin, "\"'")
	if origin == "" {
		return ""
	}
	if origin == "*" {
		return "*"
	}

	parsed, err := url.Parse(origin)
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		host := strings.ToLower(parsed.Host)
		scheme := strings.ToLower(parsed.Scheme)
		if strings.Contains(host, ":") {
			if (scheme == "https" && strings.HasSuffix(host, ":443")) || (scheme == "http" && strings.HasSuffix(host, ":80")) {
				host = strings.Split(host, ":")[0]
			}
		}
		return scheme + "://" + host
	}

	return strings.ToLower(origin)
}

func setVaryHeader(c *gin.Context, value string) {
	vary := c.Writer.Header().Get("Vary")
	if vary == "" {
		c.Writer.Header().Set("Vary", value)
		return
	}

	for _, part := range strings.Split(vary, ",") {
		if strings.EqualFold(strings.TrimSpace(part), value) {
			return
		}
	}

	c.Writer.Header().Set("Vary", vary+", "+value)
}
