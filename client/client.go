package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Kirby980/go-es/config"
	"github.com/Kirby980/go-es/errors"
	"github.com/Kirby980/go-es/logger"
)

// Client Elasticsearch 客户端
type Client struct {
	config       *config.Config
	httpClient   *http.Client
	addresses    []string
	addressIndex atomic.Uint32
	logger       logger.Logger
}

// New 创建新的 ES 客户端
func New(opts ...config.Option) (*Client, error) {
	cfg := config.DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// 初始化日志：优先使用用户传入的 Logger，否则使用默认 zap 生产日志
	var log logger.Logger
	if cfg.Logger != nil {
		log = cfg.Logger
	} else {
		zapLog, err := logger.NewDefaultLogger()
		if err != nil {
			return nil, fmt.Errorf("初始化默认日志失败: %w", err)
		}
		log = zapLog
	}

	// 配置 HTTP Transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
	}

	client := &Client{
		config:    cfg,
		addresses: cfg.Addresses,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		logger: log,
	}

	return client, nil
}

// GetLogger 返回当前客户端的日志实例（供 builder 包使用）
func (c *Client) GetLogger() logger.Logger {
	return c.logger
}

// Close 关闭客户端
func (c *Client) Close() error {
	return nil
}

// GetAddress 获取第一个地址
func (c *Client) GetAddress() string {
	if len(c.addresses) == 0 {
		return ""
	}
	idx := c.addressIndex.Add(1)
	return c.addresses[idx%uint32(len(c.addresses))]
}

// DoRequest 执行自定义 HTTP 请求
func (c *Client) DoRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	// 设置认证
	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	// 重试逻辑
	var resp *http.Response
	var err error
	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(c.config.RetryBackoff)
		}

		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}

		if c.config.EnableDebug {
			c.logger.Warn("请求失败，重试", "attempt", i+1, "maxRetries", c.config.MaxRetries, "error", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, errors.ParseESError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.addresses[0], nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("连接失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// Do 执行 HTTP 请求
func (c *Client) Do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.addresses[0] + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	// 重试逻辑
	var resp *http.Response
	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(c.config.RetryBackoff)
			// 重置 body 读取位置
			if seeker, ok := req.Body.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			}
		}

		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}

		if c.config.EnableDebug {
			c.logger.Warn("请求失败，重试", "attempt", i+1, "maxRetries", c.config.MaxRetries, "error", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, errors.ParseESError(resp.StatusCode, respBody)
	}

	return respBody, nil
}
