// Package breaker 封装出站调用的熔断逻辑。业务在调用 MySQL / Redis / 外部 HTTP
// 时，应通过 Do 包一层，避免下游故障雪崩到本服务。
//
// 每个 name 对应一个独立的熔断器实例，规则可在配置中按 name 覆盖默认值。
package breaker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sony/gobreaker/v2"

	"github.com/holosola/gorgi/internal/pkg/config"
	"github.com/holosola/gorgi/internal/pkg/errs"
)

// Manager 维护所有按 name 划分的熔断器。
type Manager struct {
	mu       sync.RWMutex
	defaults config.BreakerRule
	rules    map[string]config.BreakerRule
	cbs      map[string]*gobreaker.CircuitBreaker[any]
}

// NewManager 根据配置构造 Manager。
func NewManager(cfg config.BreakerConfig) *Manager {
	def := cfg.Default
	if def.MaxRequests == 0 {
		def.MaxRequests = 5
	}
	if def.Interval == 0 {
		def.Interval = 60 * time.Second
	}
	if def.Timeout == 0 {
		def.Timeout = 30 * time.Second
	}
	if def.ConsecutiveFailures == 0 {
		def.ConsecutiveFailures = 5
	}
	rules := make(map[string]config.BreakerRule, len(cfg.Rules))
	for k, v := range cfg.Rules {
		rules[k] = v
	}
	return &Manager{
		defaults: def,
		rules:    rules,
		cbs:      make(map[string]*gobreaker.CircuitBreaker[any]),
	}
}

// Do 在指定 name 的熔断器保护下执行 fn。
//
// 当熔断器处于 open 状态时，直接返回 errs.BreakerOpen。
func (m *Manager) Do(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	cb := m.getOrCreate(name)
	_, err := cb.Execute(func() (any, error) {
		return nil, fn(ctx)
	})
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		return errs.BreakerOpen.Wrap(err)
	}
	return err
}

// getOrCreate 懒加载某个 name 对应的熔断器实例。
func (m *Manager) getOrCreate(name string) *gobreaker.CircuitBreaker[any] {
	m.mu.RLock()
	cb, ok := m.cbs[name]
	m.mu.RUnlock()
	if ok {
		return cb
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if cb, ok := m.cbs[name]; ok {
		return cb
	}
	rule := m.defaults
	if r, ok := m.rules[name]; ok {
		mergeRule(&rule, r)
	}
	cb = gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
		Name:        name,
		MaxRequests: rule.MaxRequests,
		Interval:    rule.Interval,
		Timeout:     rule.Timeout,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.ConsecutiveFailures >= rule.ConsecutiveFailures
		},
	})
	m.cbs[name] = cb
	return cb
}

// mergeRule 用非零字段覆盖默认规则。
func mergeRule(dst *config.BreakerRule, src config.BreakerRule) {
	if src.MaxRequests != 0 {
		dst.MaxRequests = src.MaxRequests
	}
	if src.Interval != 0 {
		dst.Interval = src.Interval
	}
	if src.Timeout != 0 {
		dst.Timeout = src.Timeout
	}
	if src.ConsecutiveFailures != 0 {
		dst.ConsecutiveFailures = src.ConsecutiveFailures
	}
}
