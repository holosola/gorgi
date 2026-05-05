// Package tracing 负责 OpenTelemetry 的初始化和关闭。
//
// 默认使用 stdout 导出器（便于本地调试），后续可在配置中切换到 OTLP gRPC/HTTP。
// 当 cfg.Enabled = false 时，返回 noop tracer 与 no-op shutdown，业务零成本接入。
package tracing

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/holosola/gorgi/internal/pkg/config"
)

// ShutdownFunc 是 tracer 的关闭函数；返回的实例必须在 main 退出前调用。
type ShutdownFunc func(ctx context.Context) error

// Init 按配置初始化全局 tracer provider。
//
// 当 cfg.Enabled 为 false 时返回 no-op，调用方代码不需要任何分支判断。
func Init(cfg config.OTelConfig) (ShutdownFunc, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}
	exp, err := buildExporter(cfg)
	if err != nil {
		return nil, err
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("构建 OTel resource 失败: %w", err)
	}
	ratio := cfg.SampleRatio
	if ratio <= 0 || ratio > 1 {
		ratio = 1
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// buildExporter 根据 cfg.Exporter 选择具体的导出器，目前仅实现 stdout。
// otlp_grpc / otlp_http 留作后续扩展点。
func buildExporter(cfg config.OTelConfig) (sdktrace.SpanExporter, error) {
	switch cfg.Exporter {
	case "", "stdout":
		return stdouttrace.New(stdouttrace.WithWriter(os.Stdout), stdouttrace.WithPrettyPrint())
	case "otlp_grpc", "otlp_http":
		return nil, fmt.Errorf("OTel exporter %q 暂未实现，请使用 stdout 或自行扩展", cfg.Exporter)
	default:
		return nil, fmt.Errorf("不支持的 OTel exporter: %q", cfg.Exporter)
	}
}
