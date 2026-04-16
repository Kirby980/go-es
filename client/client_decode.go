package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Kirby980/go-es/errors"
)

// DoWithHeaderAndDecode 执行 HTTP 请求，并将响应直接流式解码到 target
func (c *Client) DoWithHeaderAndDecode(ctx context.Context, method, path string, body any, header http.Header, target any) error {
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
			return fmt.Errorf("序列化请求体失败: %w", err)
		}
	}

	// 重试逻辑
	var resp *http.Response
	var err error
	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.config.RetryBackoff):
			}
		}

		var reqBody io.Reader
		if reqBodyData != nil {
			reqBody = bytes.NewReader(reqBodyData)
		}

		// 故障转移：每次重试获取新地址
		url := c.GetAddress() + path
		req, reqErr := http.NewRequestWithContext(ctx, method, url, reqBody)
		if reqErr != nil {
			return fmt.Errorf("创建请求失败: %w", reqErr)
		}

		req.Header.Set("Content-Type", "application/json")
		for k, v := range header {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}

		if c.config.Username != "" {
			req.SetBasicAuth(c.config.Username, c.config.Password)
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
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("读取错误响应失败: %w", readErr)
		}
		return errors.ParseESError(resp.StatusCode, respBody)
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}
	return nil
}
