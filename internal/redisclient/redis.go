// Package redisclient 提供 Redis 8 客户端（go-redis/v9）。
package redisclient

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/acanx/finvquant/internal/config"
)

// NewClient 创建 Redis 客户端并 Ping 验证。
func NewClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("Redis 连通性检查失败: %w", err)
	}
	return client, nil
}
