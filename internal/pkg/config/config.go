// Package config 负责加载并管理项目运行所需的所有配置项。
//
// 加载顺序：命令行 -c 指定 > GORGI_CONFIG 环境变量 > ./configs/config.yaml。
// 加载完成后，配置会被反序列化到强类型 Config 结构上，业务代码统一通过依赖注入读取，
// 避免在各处散落 viper.GetString 调用。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/spf13/viper"
)

// Config 是项目的全局配置根。
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Log        LogConfig        `mapstructure:"log"`
	MySQL      MySQLConfig      `mapstructure:"mysql"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Middleware MiddlewareConfig `mapstructure:"middleware"`
	Breaker    BreakerConfig    `mapstructure:"breaker"`
	OTel       OTelConfig       `mapstructure:"otel"`
}

// AppConfig 是应用层级配置。
type AppConfig struct {
	Name string         `mapstructure:"name"`
	Env  string         `mapstructure:"env"`
	HTTP HTTPConfig     `mapstructure:"http"`
}

// HTTPConfig 是 HTTP 服务器配置。
type HTTPConfig struct {
	Addr         string        `mapstructure:"addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// LogConfig 是日志配置。
type LogConfig struct {
	Level      string   `mapstructure:"level"`
	Format     string   `mapstructure:"format"`
	Output     string   `mapstructure:"output"`
	FilePath   string   `mapstructure:"file_path"`
	AddSource  bool     `mapstructure:"add_source"`
	MaskFields []string `mapstructure:"mask_fields"`
}

// MySQLConfig 是 MySQL 配置。
type MySQLConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig 是 Redis 配置。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// MiddlewareConfig 汇总所有中间件相关配置。
type MiddlewareConfig struct {
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	Sign      SignConfig      `mapstructure:"sign"`
}

// CORSConfig 是跨域中间件配置。
type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
	AllowMethods []string `mapstructure:"allow_methods"`
	AllowHeaders []string `mapstructure:"allow_headers"`
}

// RateLimitConfig 是限流中间件配置。
type RateLimitConfig struct {
	Enabled bool    `mapstructure:"enabled"`
	Rate    float64 `mapstructure:"rate"`
	Burst   int     `mapstructure:"burst"`
}

// SignConfig 是验签中间件配置。
type SignConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	TimestampSkew time.Duration `mapstructure:"timestamp_skew"`
	Apps          []SignApp     `mapstructure:"apps"`
}

// SignApp 是单个接入方的签名凭证。
type SignApp struct {
	AppKey    string `mapstructure:"app_key"`
	AppSecret string `mapstructure:"app_secret"`
}

// BreakerConfig 是熔断配置。
type BreakerConfig struct {
	Default BreakerRule            `mapstructure:"default"`
	Rules   map[string]BreakerRule `mapstructure:"rules"`
}

// BreakerRule 是单条熔断规则。
type BreakerRule struct {
	MaxRequests          uint32        `mapstructure:"max_requests"`
	Interval             time.Duration `mapstructure:"interval"`
	Timeout              time.Duration `mapstructure:"timeout"`
	ConsecutiveFailures  uint32        `mapstructure:"consecutive_failures"`
}

// OTelConfig 是 OpenTelemetry 配置。
type OTelConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	ServiceName string  `mapstructure:"service_name"`
	Exporter    string  `mapstructure:"exporter"`
	SampleRatio float64 `mapstructure:"sample_ratio"`
}

// current 用 atomic.Pointer 持有当前生效的配置，热更新时整体替换。
var current atomic.Pointer[Config]

// Load 从指定路径加载配置文件。空字符串表示按默认顺序查找。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	resolved, err := resolvePath(path)
	if err != nil {
		return nil, err
	}
	v.SetConfigFile(resolved)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("加载配置文件失败 %q: %w", resolved, err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	current.Store(cfg)
	return cfg, nil
}

// MustLoad 与 Load 行为相同，加载失败时 panic，便于在 main 中使用。
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

// Current 返回最近一次加载的配置；若尚未加载则返回 nil。
func Current() *Config {
	return current.Load()
}

// resolvePath 解析配置文件路径。优先级：传入参数 > 环境变量 > 默认路径。
func resolvePath(path string) (string, error) {
	if path == "" {
		if env := os.Getenv("GORGI_CONFIG"); env != "" {
			path = env
		} else {
			path = filepath.Join("configs", "config.yaml")
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("配置文件路径无法转换为绝对路径: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("配置文件不存在: %w", err)
	}
	return abs, nil
}
