// Package api 组装 Gin 路由。
package api

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/acanx/finvquant/internal/api/handler"
	"github.com/acanx/finvquant/internal/quote"
)

// Deps 路由依赖。
type Deps struct {
	Pool    *pgxpool.Pool
	Redis   *redis.Client
	Version string
	Commit  string
	Server  string
}

// NewRouter 构建 Gin 路由（默认监听端口 16001）。
func NewRouter(deps *Deps) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	health := &handler.Health{
		Pool:   deps.Pool,
		Redis:  deps.Redis,
		Server: deps.Server,
	}
	version := &handler.Version{
		Version:   deps.Version,
		GoVersion: runtime.Version(),
		Commit:    deps.Commit,
	}
	quoteImport := handler.NewQuoteImport(quote.NewService(deps.Pool))

	apiGroup := router.Group("/API/V1")
	{
		apiGroup.GET("/health/live", health.Live)
		apiGroup.GET("/health/ready", health.Ready)
		apiGroup.GET("/version", version.Info)
		apiGroup.POST("/Quote/Import/Upload", quoteImport.Upload)
	}

	return router
}
