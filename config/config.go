package config

import (
	"time"
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
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addresses:     []string{"http://localhost:9200"},
		MaxRetries:    3,
		RetryBackoff:  time.Second,
		Timeout:       30 * time.Second,
		EnableMetrics: false,
		EnableDebug:   false,
	}
}

// Option 配置选项函数
type Option func(*Config)

// WithTransport 设置传输层
func WithTransport(skip bool) Option {
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
