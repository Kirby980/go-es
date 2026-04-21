package config

import (
	"context"
	"net/http"
	"time"

	"github.com/Kirby980/go-es/logger"
)

// Config Elasticsearch 配置
type Config struct {
	// ES 服务器地址列表
	Addresses []string

	// 认证信息
	Username string
	Password string

	// 连接配置
	MaxRetries   int
	RetryBackoff time.Duration

	// 跳过证书验证
	InsecureSkipVerify bool

	// 超时配置
	Timeout time.Duration

	// 其他配置
	EnableMetrics bool
	EnableDebug   bool
	EnableGzip    bool

	EnableExponentialBackoff bool
	MaxRetryBackoff          time.Duration

	EnableCircuitBreaker      bool
	CircuitBreakerFailures    int
	CircuitBreakerCooldown    time.Duration
	CircuitBreakerHealthCheck time.Duration

	EnableSniff   bool
	SniffInterval time.Duration

	// Logger 自定义日志实现，nil 时 client.New() 使用默认 zap 生产日志
	Logger logger.Logger
	// 连接池配置
	MaxIdleConns        int           // 最大空闲连接
	MaxIdleConnsPerHost int           // 每个主机的最大空闲连接数
	MaxConnsPerHost     int           // 每个主机的最大连接数
	IdleConnTimeout     time.Duration // 空闲连接超时
	Hooks               []Hook
}

// Hook 定义了客户端请求的生命周期钩子，用于实现可观测性（Tracing/Metrics）
type Hook interface {
	BeforeRequest(ctx context.Context, req *http.Request) context.Context
	AfterRequest(ctx context.Context, req *http.Request, resp *http.Response, duration time.Duration)
	OnError(ctx context.Context, req *http.Request, err error, duration time.Duration)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addresses:                 []string{"http://localhost:9200"},
		MaxRetries:                3,
		RetryBackoff:              time.Second,
		Timeout:                   30 * time.Second,
		EnableMetrics:             false,
		EnableDebug:               false,
		EnableGzip:                false,
		EnableExponentialBackoff:  false,
		MaxRetryBackoff:           30 * time.Second,
		EnableCircuitBreaker:      false,
		CircuitBreakerFailures:    3,
		CircuitBreakerCooldown:    10 * time.Second,
		CircuitBreakerHealthCheck: 5 * time.Second,
		EnableSniff:               false,
		SniffInterval:             5 * time.Minute,
		MaxIdleConns:              100,
		MaxIdleConnsPerHost:       10,
		MaxConnsPerHost:           0,
		IdleConnTimeout:           90 * time.Second,
	}
}

// Option 配置选项函数
type Option func(*Config)

// WithInsecureSkipVerify 设置传输层
func WithInsecureSkipVerify(skip bool) Option {
	return func(c *Config) {
		c.InsecureSkipVerify = skip
	}
}

// WithAddresses 设置 ES 地址
func WithAddresses(addresses ...string) Option {
	return func(c *Config) {
		c.Addresses = addresses
	}
}

// WithAuth 设置认证信息
func WithAuth(username, password string) Option {
	return func(c *Config) {
		c.Username = username
		c.Password = password
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetry 设置重试配置
func WithRetry(maxRetries int, backoff time.Duration) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
		c.RetryBackoff = backoff
	}
}

// WithDebug 启用调试模式
func WithDebug(enable bool) Option {
	return func(c *Config) {
		c.EnableDebug = enable
	}
}

// WithGzip ...
func WithGzip(enable bool) Option {
	return func(c *Config) {
		c.EnableGzip = enable
	}
}

// WithExponentialBackoff ...
func WithExponentialBackoff(enable bool, maxBackoff time.Duration) Option {
	return func(c *Config) {
		c.EnableExponentialBackoff = enable
		if maxBackoff > 0 {
			c.MaxRetryBackoff = maxBackoff
		}
	}
}

// WithCircuitBreaker ...
func WithCircuitBreaker(enable bool, failures int, cooldown time.Duration, healthCheck time.Duration) Option {
	return func(c *Config) {
		c.EnableCircuitBreaker = enable
		if failures > 0 {
			c.CircuitBreakerFailures = failures
		}
		if cooldown > 0 {
			c.CircuitBreakerCooldown = cooldown
		}
		if healthCheck > 0 {
			c.CircuitBreakerHealthCheck = healthCheck
		}
	}
}

// WithSniff ...
func WithSniff(enable bool, interval time.Duration) Option {
	return func(c *Config) {
		c.EnableSniff = enable
		if interval > 0 {
			c.SniffInterval = interval
		}
	}
}

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *Config) {
		c.MaxIdleConns = maxIdleConns
	}
}

// WithMaxIdleConnsPerHost 设置每个主机的最大空闲连接数
func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return func(c *Config) {
		c.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

// WithMaxConnsPerHost  设置每个host最大连接数
func WithMaxConnsPerHost(maxConnsPerHost int) Option {
	return func(c *Config) {
		c.MaxConnsPerHost = maxConnsPerHost
	}
}

// WithIdleConnTimeout 设置空闲连接超时时间
func WithIdleConnTimeout(idleConnTimeout time.Duration) Option {
	return func(c *Config) {
		c.IdleConnTimeout = idleConnTimeout
	}
}

// WithLogger 设置自定义日志实现
// 传入 logger.NopLogger{} 可完全禁用日志输出
// 示例：
//
//	zapDev, _ := logger.NewDevelopmentLogger()
//	client.New(config.WithLogger(zapDev))
func WithLogger(l logger.Logger) Option {
	return func(c *Config) {
		c.Logger = l
	}
}

// WithHooks 设置客户端请求的生命周期钩子（用于实现 Tracing/Metrics 等可观测性）
func WithHooks(hooks ...Hook) Option {
	return func(c *Config) {
		c.Hooks = append(c.Hooks, hooks...)
	}
}
