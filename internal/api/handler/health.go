// Package handler 提供 HTTP 处理器。
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Health 健康检查处理器（依赖 PG 与 Redis 探针）。
type Health struct {
	Pool   *pgxpool.Pool
	Redis  *redis.Client
	Server string
}

// Live 存活探针：进程存活即 200。
func (h *Health) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"server": h.Server,
	})
}

// Ready 就绪探针：PG 与 Redis 均可达才 200，否则 503。
func (h *Health) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	status := "ok"
	code := http.StatusOK
	details := gin.H{}

	if err := h.Pool.Ping(ctx); err != nil {
		status = "degraded"
		code = http.StatusServiceUnavailable
		details["postgres"] = "unreachable"
	} else {
		details["postgres"] = "ok"
	}

	if err := h.Redis.Ping(ctx).Err(); err != nil {
		status = "degraded"
		code = http.StatusServiceUnavailable
		details["redis"] = "unreachable"
	} else {
		details["redis"] = "ok"
	}

	c.JSON(code, gin.H{
		"status":  status,
		"server":  h.Server,
		"details": details,
	})
}

// Version 版本信息。
type Version struct {
	Version   string
	GoVersion string
	Commit    string
}

// Info 返回服务端版本信息。
func (v *Version) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":       "finvquant-server",
		"version":    v.Version,
		"go_version": v.GoVersion,
		"commit":     v.Commit,
	})
}
