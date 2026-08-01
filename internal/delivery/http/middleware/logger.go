package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RequestLogger middleware logs HTTP requests using zerolog.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		var event *zerolog.Event
		if statusCode >= 500 {
			event = log.Error()
		} else if statusCode >= 400 {
			event = log.Warn()
		} else {
			event = log.Info()
		}

		event.
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Str("ip", clientIP).
			Dur("latency_ms", latency).
			Msg("HTTP Request")
	}
}
