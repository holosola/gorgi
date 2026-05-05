// Package server 装配并运行 HTTP 服务，封装优雅退出逻辑。
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/pkg/config"
	"github.com/holosola/gorgi/internal/pkg/log"
)

// New 根据配置创建 *http.Server。
func New(cfg config.HTTPConfig, handler *gin.Engine) *http.Server {
	addr := cfg.Addr
	if addr == "" {
		addr = ":8080"
	}
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  defaultDuration(cfg.ReadTimeout, 60*time.Second),
		WriteTimeout: defaultDuration(cfg.WriteTimeout, 60*time.Second),
		IdleTimeout:  defaultDuration(cfg.IdleTimeout, 120*time.Second),
	}
}

// RunWithGracefulShutdown 在阻塞模式下启动 HTTP 服务，并监听 SIGINT/SIGTERM。
//
// 收到信号后：先 Shutdown HTTP server（10 秒超时），再调用 onShutdown 释放外部依赖。
// onShutdown 可以为 nil。
func RunWithGracefulShutdown(srv *http.Server, onShutdown func(ctx context.Context)) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		log.L(ctx).Info("HTTP 服务启动", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("HTTP 服务异常退出: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.L(ctx).Info("收到退出信号，开始优雅关闭")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.L(shutdownCtx).Error("HTTP server 关闭失败", "err", err.Error())
	}
	if onShutdown != nil {
		onShutdown(shutdownCtx)
	}
	if signalOSExit != nil {
		// 仅测试时注入，正常路径为 nil。
		signalOSExit()
	}
	return nil
}

// signalOSExit 仅测试用钩子。
var signalOSExit func()

// defaultDuration 当 v <= 0 时返回 def。
func defaultDuration(v, def time.Duration) time.Duration {
	if v <= 0 {
		return def
	}
	return v
}

// 兼容跨平台的 SIGTERM 引用：在 Windows 上没有 SIGTERM，但 syscall 仍提供常量。
// 这里把 os.Interrupt 也囊括进去（部分容器只发 SIGINT）。
var _ = os.Interrupt
