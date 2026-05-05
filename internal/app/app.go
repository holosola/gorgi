// Package app 是应用装配层：把配置、日志、tracing、数据库、路由等串起来。
//
// 设计取舍：MySQL / Redis 在 DSN（或 addr）为空时跳过初始化，方便本地无依赖启动整个框架；
// 生产配置必须填写完整。其他可选依赖同理。
package app

import (
	"context"
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"

	"github.com/holosola/gorgi/internal/app/router"
	"github.com/holosola/gorgi/internal/app/server"
	"github.com/holosola/gorgi/internal/pkg/breaker"
	"github.com/holosola/gorgi/internal/pkg/config"
	"github.com/holosola/gorgi/internal/pkg/log"
	"github.com/holosola/gorgi/internal/pkg/mysql"
	"github.com/holosola/gorgi/internal/pkg/redis"
	"github.com/holosola/gorgi/internal/pkg/tracing"
)

// Run 是应用启动入口。cfgPath 传空则按 config 包的默认顺序查找。
func Run(cfgPath string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	if err := log.Init(cfg.Log); err != nil {
		return err
	}
	if cfg.App.Env != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	shutdownTracer, err := tracing.Init(cfg.OTel)
	if err != nil {
		return err
	}

	db, rdb, err := initStores(ctx, cfg)
	if err != nil {
		return err
	}

	deps := router.Deps{
		Cfg:     cfg,
		Breaker: breaker.NewManager(cfg.Breaker),
	}
	engine := router.New(deps, modules(deps)...)
	srv := server.New(cfg.App.HTTP, engine)

	return server.RunWithGracefulShutdown(srv, func(shutdownCtx context.Context) {
		_ = shutdownTracer(shutdownCtx)
		closeStores(shutdownCtx, db, rdb)
	})
}

// initStores 初始化 MySQL 与 Redis；DSN/Addr 为空表示跳过（仅本地开发可用）。
func initStores(ctx context.Context, cfg *config.Config) (*sql.DB, *goredis.Client, error) {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var db *sql.DB
	if cfg.MySQL.DSN != "" {
		var err error
		db, err = mysql.New(pingCtx, cfg.MySQL)
		if err != nil {
			return nil, nil, err
		}
	} else {
		log.L(ctx).Warn("mysql.dsn 为空，跳过 MySQL 初始化")
	}

	var rdb *goredis.Client
	if cfg.Redis.Addr != "" {
		var err error
		rdb, err = redis.New(pingCtx, cfg.Redis)
		if err != nil {
			if db != nil {
				_ = db.Close()
			}
			return nil, nil, err
		}
	} else {
		log.L(ctx).Warn("redis.addr 为空，跳过 Redis 初始化")
	}
	return db, rdb, nil
}

// closeStores 关闭 MySQL 与 Redis 连接。
func closeStores(_ context.Context, db *sql.DB, rdb *goredis.Client) {
	if db != nil {
		_ = db.Close()
	}
	if rdb != nil {
		_ = rdb.Close()
	}
}
