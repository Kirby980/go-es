package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Kirby980/go-es/config"
	"github.com/Kirby980/go-es/errors"
	"github.com/Kirby980/go-es/logger"
)

type hookRoundTripper struct {
	rt    http.RoundTripper
	hooks []config.Hook
}

func (h *hookRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	for _, hook := range h.hooks {
		ctx = hook.BeforeRequest(ctx, req)
	}
	req = req.WithContext(ctx)

	start := time.Now()
	resp, err := h.rt.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		for _, hook := range h.hooks {
			hook.OnError(ctx, req, err, duration)
		}
	} else {
		for _, hook := range h.hooks {
			hook.AfterRequest(ctx, req, resp, duration)
		}
	}

	return resp, err
}

// Client Elasticsearch 客户端
type Client struct {
	config       *config.Config
	httpClient   *http.Client
	addresses    []string
	addressIndex atomic.Uint32
	logger       logger.Logger
	breakerMu    sync.Mutex
	breakerState map[string]*nodeState
	breakerStop  chan struct{}
}

type nodeState struct {
	failures       int
	unhealthyUntil time.Time
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
	var transport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
	}

	if len(cfg.Hooks) > 0 {
		transport = &hookRoundTripper{
			rt:    transport,
			hooks: cfg.Hooks,
		}
	}

	client := &Client{
		config:    cfg,
		addresses: cfg.Addresses,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		logger:       log,
		breakerState: make(map[string]*nodeState),
	}

	if cfg.EnableCircuitBreaker {
		client.breakerStop = make(chan struct{})
		for _, addr := range cfg.Addresses {
			client.breakerState[addr] = &nodeState{}
		}
		go client.runHealthCheck()
	}

	return client, nil
}

// GetLogger 返回当前客户端的日志实例（供 builder 包使用）
func (c *Client) GetLogger() logger.Logger {
	return c.logger
}

// Close 关闭客户端
func (c *Client) Close() error {
	if c.breakerStop != nil {
		close(c.breakerStop)
	}
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// GetAddress 获取第一个地址
func (c *Client) GetAddress() string {
	if len(c.addresses) == 0 {
		return ""
	}
	idx := c.addressIndex.Add(1) - 1
	return c.addresses[idx%uint32(len(c.addresses))]
}

func (c *Client) getHealthyAddress() string {
	if len(c.addresses) == 0 {
		return ""
	}
	if !c.config.EnableCircuitBreaker {
		return c.GetAddress()
	}
	start := int(c.addressIndex.Add(1)-1) % len(c.addresses)
	now := time.Now()

	c.breakerMu.Lock()
	defer c.breakerMu.Unlock()

	for i := 0; i < len(c.addresses); i++ {
		addr := c.addresses[(start+i)%len(c.addresses)]
		st := c.breakerState[addr]
		if st == nil {
			st = &nodeState{}
			c.breakerState[addr] = st
		}
		if st.unhealthyUntil.IsZero() || st.unhealthyUntil.Before(now) {
			return addr
		}
	}
	return c.addresses[start]
}

func (c *Client) markFailure(addr string) {
	if !c.config.EnableCircuitBreaker || addr == "" {
		return
	}
	c.breakerMu.Lock()
	defer c.breakerMu.Unlock()
	st := c.breakerState[addr]
	if st == nil {
		st = &nodeState{}
		c.breakerState[addr] = st
	}
	st.failures++
	if st.failures >= c.config.CircuitBreakerFailures {
		st.unhealthyUntil = time.Now().Add(c.config.CircuitBreakerCooldown)
		st.failures = 0
	}
}

func (c *Client) markSuccess(addr string) {
	if !c.config.EnableCircuitBreaker || addr == "" {
		return
	}
	c.breakerMu.Lock()
	defer c.breakerMu.Unlock()
	st := c.breakerState[addr]
	if st == nil {
		st = &nodeState{}
		c.breakerState[addr] = st
	}
	st.failures = 0
	st.unhealthyUntil = time.Time{}
}

func (c *Client) shouldRetryStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return true
	default:
		return code >= 500
	}
}

func (c *Client) backoffForAttempt(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	base := c.config.RetryBackoff
	if !c.config.EnableExponentialBackoff {
		return base
	}
	maxB := c.config.MaxRetryBackoff
	if maxB <= 0 {
		maxB = 30 * time.Second
	}
	d := base * time.Duration(1<<uint(attempt-1))
	if d > maxB {
		return maxB
	}
	return d
}

func (c *Client) runHealthCheck() {
	ticker := time.NewTicker(c.config.CircuitBreakerHealthCheck)
	defer ticker.Stop()

	for {
		select {
		case <-c.breakerStop:
			return
		case <-ticker.C:
		}

		addrs := c.addresses
		for _, addr := range addrs {
			c.breakerMu.Lock()
			st := c.breakerState[addr]
			unhealthy := st != nil && !st.unhealthyUntil.IsZero() && st.unhealthyUntil.After(time.Now())
			c.breakerMu.Unlock()
			if !unhealthy {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
			if err != nil {
				cancel()
				continue
			}
			if c.config.Username != "" {
				req.SetBasicAuth(c.config.Username, c.config.Password)
			}
			resp, err := c.httpClient.Do(req)
			cancel()
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				c.markSuccess(addr)
			}
		}
	}
}

func (c *Client) gzipBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
	var gzBody []byte

	if c.config.EnableGzip && req.Body != nil && req.Header.Get("Content-Encoding") == "" {
		raw, readErr := io.ReadAll(req.Body)
		if readErr != nil {
			return nil, fmt.Errorf("读取请求体失败: %w", readErr)
		}
		req.Body.Close()

		gzBody, err = c.gzipBytes(raw)
		if err != nil {
			return nil, fmt.Errorf("压缩请求体失败: %w", err)
		}
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(gzBody)), nil
		}
		req.Body, _ = req.GetBody()
	}

	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.backoffForAttempt(i)):
			}

			// 重置 body 读取位置，防止重试时发送空 body
			if req.GetBody != nil {
				req.Body, _ = req.GetBody()
			} else if req.Body != nil {
				if seeker, ok := req.Body.(io.Seeker); ok {
					seeker.Seek(0, io.SeekStart)
				}
			}

			// 故障转移：切换到新地址
			newAddr := c.getHealthyAddress()
			if newAddr != "" {
				if parsedURL, parseErr := url.Parse(newAddr); parseErr == nil {
					req.URL.Scheme = parsedURL.Scheme
					req.URL.Host = parsedURL.Host
					req.Host = parsedURL.Host
				}
			}
		}

		addrKey := req.URL.Scheme + "://" + req.URL.Host
		resp, err = c.httpClient.Do(req)
		if err == nil && (resp == nil || !c.shouldRetryStatus(resp.StatusCode)) {
			c.markSuccess(addrKey)
			break
		}

		c.markFailure(addrKey)
		if c.config.EnableDebug {
			c.logger.Warn("请求失败，重试", "attempt", i+1, "maxRetries", c.config.MaxRetries, "error", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respReader := io.Reader(resp.Body)
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gr, gzErr := gzip.NewReader(resp.Body)
		if gzErr != nil {
			return nil, fmt.Errorf("解压响应失败: %w", gzErr)
		}
		defer gr.Close()
		respReader = gr
	}

	respBody, err := io.ReadAll(respReader)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, errors.ParseESError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// Ping 测试连接（遍历所有地址，任一可达即返回 nil）
func (c *Client) Ping(ctx context.Context) error {
	if len(c.addresses) == 0 {
		return fmt.Errorf("no addresses configured")
	}

	var lastErr error
	for _, addr := range c.addresses {
		// 每次循环检查 context 是否已取消，以便及时退出
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("连接失败: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", addr, nil)
		if err != nil {
			lastErr = fmt.Errorf("创建请求失败: %w", err)
			continue
		}

		if c.config.Username != "" {
			req.SetBasicAuth(c.config.Username, c.config.Password)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("连接失败: %w", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("连接失败，状态码: %d", resp.StatusCode)
			continue
		}

		// At least one address responded successfully.
		return nil
	}

	return lastErr
}

// Do 执行 HTTP 请求
func (c *Client) Do(ctx context.Context, method, path string, body any) ([]byte, error) {
	return c.DoWithHeader(ctx, method, path, body, nil)
}

// DoWithHeader 执行 HTTP 请求并带上自定义 Header
func (c *Client) DoWithHeader(ctx context.Context, method, path string, body any, header http.Header) ([]byte, error) {
	// 如果 context 没有设置截止时间，应用默认超时
	if _, ok := ctx.Deadline(); !ok && c.httpClient.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.httpClient.Timeout)
		defer cancel()
	}

	var reqBodyData []byte
	if body != nil {
		var err error
		reqBodyData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
	}

	// 重试逻辑
	var resp *http.Response
	var err error
	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.backoffForAttempt(i)):
			}
		}

		// 故障转移：每次重试获取新地址
		addr := c.getHealthyAddress()
		url := addr + path

		var reqBody io.Reader
		if reqBodyData != nil {
			if c.config.EnableGzip {
				gzData, gzErr := c.gzipBytes(reqBodyData)
				if gzErr != nil {
					return nil, fmt.Errorf("压缩请求体失败: %w", gzErr)
				}
				reqBody = bytes.NewReader(gzData)
			} else {
				reqBody = bytes.NewReader(reqBodyData)
			}
		}

		req, reqErr := http.NewRequestWithContext(ctx, method, url, reqBody)
		if reqErr != nil {
			return nil, fmt.Errorf("创建请求失败: %w", reqErr)
		}

		req.Header.Set("Content-Type", "application/json")
		if c.config.EnableGzip && reqBodyData != nil {
			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Accept-Encoding", "gzip")
		} else if c.config.EnableGzip {
			req.Header.Set("Accept-Encoding", "gzip")
		}
		for k, v := range header {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}

		if c.config.Username != "" {
			req.SetBasicAuth(c.config.Username, c.config.Password)
		}

		resp, err = c.httpClient.Do(req)
		if err == nil && (resp == nil || !c.shouldRetryStatus(resp.StatusCode)) {
			c.markSuccess(addr)
			break
		}

		c.markFailure(addr)
		if c.config.EnableDebug {
			c.logger.Warn("请求失败，重试", "attempt", i+1, "maxRetries", c.config.MaxRetries, "error", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respReader := io.Reader(resp.Body)
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gr, gzErr := gzip.NewReader(resp.Body)
		if gzErr != nil {
			return nil, fmt.Errorf("解压响应失败: %w", gzErr)
		}
		defer gr.Close()
		respReader = gr
	}

	respBody, err := io.ReadAll(respReader)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, errors.ParseESError(resp.StatusCode, respBody)
	}

	return respBody, nil
}
