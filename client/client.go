package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Kirby980/go-es/config"
	"github.com/Kirby980/go-es/errors"
)

// Client Elasticsearch 客户端
type Client struct {
	config     *config.Config
	httpClient *http.Client
	addresses  []string
}

// New 创建新的 ES 客户端
func New(opts ...config.Option) (*Client, error) {
	cfg := config.DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
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
	}

	return client, nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	return nil
}

// GetAddress 获取第一个地址
func (c *Client) GetAddress() string {
	if len(c.addresses) > 0 {
		return c.addresses[0]
	}
	return ""
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
			fmt.Printf("请求失败，重试 %d/%d: %v\n", i+1, c.config.MaxRetries, err)
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
func (c *Client) Do(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
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
		}

		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}

		if c.config.EnableDebug {
			fmt.Printf("请求失败，重试 %d/%d: %v\n", i+1, c.config.MaxRetries, err)
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
