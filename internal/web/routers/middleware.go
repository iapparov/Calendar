package routers

import (
	"calendar/internal/logger"
	"calendar/internal/web/handlers"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func LoggerMiddleware(l *logger.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		l.Log(
			zapcore.InfoLevel,
			"HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
		)
	}
}

func AuthMiddleware(auth handlers.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errAuthMissing)
			return
		}

		// Case-insensitive "Bearer " prefix check without allocating a lowered copy.
		const bearer = "Bearer "
		if len(authHeader) < len(bearer) ||
			!strings.EqualFold(authHeader[:len(bearer)], bearer) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errAuthScheme)
			return
		}

		token := strings.TrimSpace(authHeader[len(bearer):])
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errTokenEmpty)
			return
		}

		payload, err := auth.ValidateTokens(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errTokenInvalid)
			return
		}

		c.Set(handlers.CtxUserID, payload.UserID)
		c.Next()
	}
}
