package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

const (
	requestIDContextKey = "requestID"
	loggerContextKey    = "logger"
	requestIDHeader     = "X-Request-ID"
	requestIDField      = "request_id"
)

func GetLoggerFromContext(ctx *gin.Context) *zap.Logger {
	if ctx == nil {
		return zap.L().With(zap.String("error", "nil_context"))
	}

	if logger, exists := ctx.Get(loggerContextKey); exists {
		zapLogger, ok := logger.(*zap.Logger)
		if !ok {
			return zap.L().With(zap.String("error", "invalid_logger_type"))
		}
		return zapLogger
	}
	return zap.L().With(zap.String("error", "logger_not_found_in_context"))
}
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = xid.New().String()
		}
		c.Writer.Header().Set(requestIDHeader, requestID)

		requestLogger := logger.With(
			zap.String(requestIDField, requestID),
		)
		c.Set(loggerContextKey, requestLogger)
		c.Set(requestIDContextKey, requestID)

		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		requestLogger.Info("request started",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
		)

		c.Next()

		end := time.Now().UTC()
		latency := end.Sub(start)

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
		}

		if c.Writer.Status() >= 500 {
			requestLogger.Error("request completed", fields...)
		} else if c.Writer.Status() >= 400 {
			requestLogger.Warn("request completed", fields...)
		} else {
			requestLogger.Info("request completed", fields...)
		}
	}
}
