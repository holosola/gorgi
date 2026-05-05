package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/holosola/gorgi/internal/pkg/config"
	"github.com/holosola/gorgi/internal/pkg/errs"
	"github.com/holosola/gorgi/internal/pkg/response"
)

// ipLimiter 维护按 IP 维度的令牌桶。简化实现没有做 LRU 淘汰，
// 长期运行的服务建议把它替换成带容量上限的 LRU（如 hashicorp/golang-lru）。
type ipLimiter struct {
	mu      sync.Mutex
	rate    rate.Limit
	burst   int
	buckets map[string]*rate.Limiter
}

// newIPLimiter 构造 ipLimiter。
func newIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{rate: r, burst: b, buckets: make(map[string]*rate.Limiter)}
}

// limiterFor 返回（或懒创建）某个 IP 的 limiter。
func (l *ipLimiter) limiterFor(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.buckets[ip]
	if !ok {
		lim = rate.NewLimiter(l.rate, l.burst)
		l.buckets[ip] = lim
	}
	return lim
}

// RateLimit 按 IP 进行入站限流。enabled=false 时直接放行。
func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	if !cfg.Enabled || cfg.Rate <= 0 || cfg.Burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	limiter := newIPLimiter(rate.Limit(cfg.Rate), cfg.Burst)
	return func(c *gin.Context) {
		if !limiter.limiterFor(c.ClientIP()).Allow() {
			response.AbortFail(c, errs.RateLimited)
			return
		}
		c.Next()
	}
}
