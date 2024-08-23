package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whoisnian/tracing-benchmark/server/global"
)

func pingRawHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func pingRedisHandler(c *gin.Context) {
	err := global.RDB.Ping(c.Request.Context()).Err()
	if err != nil {
		global.LOG.ErrorContext(c.Request.Context(), "redis ping", slog.Any("error", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.String(http.StatusOK, "pong")
}

func pingMysqlHandler(c *gin.Context) {
	err := global.DB.Raw("SELECT 1").Error
	if err != nil {
		global.LOG.ErrorContext(c.Request.Context(), "mysql select 1", slog.Any("error", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.String(http.StatusOK, "pong")
}
