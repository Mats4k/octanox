package octanox

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func logger() gin.HandlerFunc {
	return gin.Logger()
}

func cors() gin.HandlerFunc {
	corsAllowedOrigin := os.Getenv("NOX__CORS_ALLOWED_ORIGINS")

	return func(c *gin.Context) {
		if corsAllowedOrigin == "*" {
			requestDomain := c.Request.Header.Get("Origin")
			c.Writer.Header().Set("Access-Control-Allow-Origin", requestDomain)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", corsAllowedOrigin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Baggage, Accept, Sentry-Trace")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				failedReq, ok := err.(failedRequest)
				if ok {
					c.JSON(failedReq.status, gin.H{"error": failedReq.message})
					return
				}

				Current.emitError(Error(fmt.Errorf("internal REST Server Error: %v", err)))

				c.JSON(500, gin.H{"error": "Internal Server Error"})
			}
		}()
		c.Next()
	}
}

// errorCollectorToHandler emits all collected errors in the Gin context to the error handlers.
func errorCollectorToHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			Current.emitError(fmt.Errorf("gin error: %s", c.Errors.String()))
		}
	}
}
