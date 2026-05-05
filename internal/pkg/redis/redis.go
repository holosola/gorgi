// Package redis 封装 go-redis 客户端的初始化。
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/holosola/gorgi/internal/pkg/config"
)

// New 创建并 ping Redis 客户端。
func New(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("redis.addr 未配置")
	}
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	if err := cli.Ping(ctx).Err(); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("ping Redis 失败: %w", err)
	}
	return cli, nil
}
