package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func InitLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// StructuredLogger เป็น Middleware สำหรับ Gin
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		//ให้ Request วิ่งไปทำงานต่อใน Handler ถัดไป
		c.Next()
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		attrs := []slog.Attr{
			slog.Int("status", statusCode),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency_ns", latency),
			slog.String("latency_human", latency.String()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		if errorMessage != "" {
			attrs = append(attrs, slog.String("error", errorMessage))
		}

		// แยก Log Level ตาม HTTP Status Cod
		msg := "incoming request"
		if statusCode >= 500 {
			slog.LogAttrs(c.Request.Context(), slog.LevelError, msg, attrs...)
		} else if statusCode >= 400 {
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, msg, attrs...)
		} else {
			slog.LogAttrs(c.Request.Context(), slog.LevelInfo, msg, attrs...)
		}
	}
}
