package router

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whoisnian/tracing-benchmark/global"
)

func pingGinHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func pingGinRedisHandler(c *gin.Context) {
	if err := pingRedis(c.Request.Context()); err != nil {
		global.LOG.ErrorContext(c.Request.Context(), "ping redis", slog.Any("error", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.String(http.StatusOK, "pong")
}

func pingGinMysqlHandler(c *gin.Context) {
	if err := pingMysql(c.Request.Context()); err != nil {
		global.LOG.ErrorContext(c.Request.Context(), "ping mysql", slog.Any("error", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.String(http.StatusOK, "pong")
}

func pingGinRedisMysqlHandler(c *gin.Context) {
	if err := pingRedis(c.Request.Context()); err != nil {
		global.LOG.ErrorContext(c.Request.Context(), "ping redis", slog.Any("error", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if err := pingMysql(c.Request.Context()); err != nil {
		global.LOG.ErrorContext(c.Request.Context(), "ping mysql", slog.Any("error", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.String(http.StatusOK, "pong")
}

func pingRedis(ctx context.Context) error {
	return global.RDB.Ping(ctx).Err()
}

func pingMysql(ctx context.Context) error {
	return global.DB.WithContext(ctx).Exec("SELECT 1").Error
}
