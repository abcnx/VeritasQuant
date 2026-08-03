// Package server 提供 FinvQuant 量化策略交易平台服务端入口。
//
// 技术栈：Go 1.25.3 + Gin v1.12 + PostgreSQL 18（pgx/v5）+ Redis 8（go-redis/v9）。
// 默认监听端口：16001。
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/acanx/finvquant/internal/api"
	"github.com/acanx/finvquant/internal/config"
	"github.com/acanx/finvquant/internal/database"
	"github.com/acanx/finvquant/internal/redisclient"
)

// 版本信息（构建时通过 -ldflags 注入）。
var (
	version = "0.1.0"
	commit  = "dev"
)

func main() {
	cfg := config.Load()
	if cfg.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// PG 18 连接池
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("[finvquant] %v", err)
	}
	defer pool.Close()

	// Redis 8 客户端
	rdb, err := redisclient.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("[finvquant] %v", err)
	}
	defer rdb.Close()

	router := api.NewRouter(&api.Deps{
		Pool:    pool,
		Redis:   rdb,
		Version: version,
		Commit:  commit,
		Server:  "finvquant-server",
	})

	server := &http.Server{
		Addr:              cfg.ListenAddr(),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[finvquant] 服务端启动，监听 %s（Go %s）", cfg.ListenAddr(), runtime.Version())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[finvquant] 服务端异常退出: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("[finvquant] 收到退出信号，优雅关闭...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
