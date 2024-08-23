package router

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoisnian/tracing-benchmark/server/global"
)

func Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.RouterGroup.Use(Logger(global.LOG))
	engine.RouterGroup.Use(Recovery(global.LOG))
	engine.NoRoute()
	engine.NoMethod()

	engine.Handle(http.MethodGet, "/ping/raw", pingRawHandler)
	engine.Handle(http.MethodGet, "/ping/redis", pingRedisHandler)

	return engine
}

// https://github.com/gin-gonic/gin/blob/75ccf94d605a05fe24817fc2f166f6f2959d5cea/logger.go#L212
// https://github.com/gin-contrib/zap/blob/4d85a5c57393196a4e19c984bc5b17a63f936710/zap.go#L57
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		fields := []any{
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.String("user-agent", c.Request.UserAgent()),
			slog.Duration("latency", time.Since(start)),
		}

		if len(c.Errors) > 0 {
			logger.ErrorContext(c.Request.Context(), c.Errors.String(), fields...)
		} else {
			logger.InfoContext(c.Request.Context(), "", fields...)
		}
	}
}

// https://github.com/gin-gonic/gin/blob/75ccf94d605a05fe24817fc2f166f6f2959d5cea/recovery.go#L51
// https://github.com/gin-contrib/zap/blob/4d85a5c57393196a4e19c984bc5b17a63f936710/zap.go#L146
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.ErrorContext(c.Request.Context(), c.Request.URL.Path,
						slog.Any("error", err),
						slog.String("request", string(httpRequest)),
					)
					c.Error(err.(error))
					c.Abort()
					return
				}

				logger.ErrorContext(c.Request.Context(), "[Recovery from panic]",
					slog.Any("error", err),
					slog.String("request", string(httpRequest)),
					slog.String("stack", string(debug.Stack())),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
