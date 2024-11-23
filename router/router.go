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
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/whoisnian/tracing-benchmark/global"
	"go.elastic.co/apm/module/apmgin/v2"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	switch global.CFG.TraceBackend {
	case "otlp":
		engine.RouterGroup.Use(otelgin.Middleware(""))
	case "apm":
		engine.RouterGroup.Use(apmgin.Middleware(engine))
	case "zipkin":
		wrapH := zipkinhttp.NewServerMiddleware(global.TR.Source().(*zipkin.Tracer))
		engine.RouterGroup.Use(func(c *gin.Context) {
			next := func(_ http.ResponseWriter, _ *http.Request) { c.Next() }
			wrapH(http.HandlerFunc(next)).ServeHTTP(c.Writer, c.Request)
		})
	}
	engine.RouterGroup.Use(Metrics(global.MT))
	engine.RouterGroup.Use(Logger(global.LOG))
	engine.RouterGroup.Use(Recovery(global.LOG))
	engine.NoRoute()
	engine.NoMethod()

	engine.Handle(http.MethodGet, "/ping/G", pingGinHandler)             // trace: gin
	engine.Handle(http.MethodGet, "/ping/GR", pingGinRedisHandler)       // trace: gin + redis
	engine.Handle(http.MethodGet, "/ping/GM", pingGinMysqlHandler)       // trace: gin + mysql
	engine.Handle(http.MethodGet, "/ping/GRM", pingGinRedisMysqlHandler) // trace: gin + redis + mysql
	engine.Handle(http.MethodGet, "/metrics", gin.WrapH(global.MT.Handler))

	return engine
}

func Metrics(mt *global.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		mt.RecordRequest(c.Writer.Status())
	}
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
