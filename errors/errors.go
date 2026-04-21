package errors

import (
	"encoding/json"
	"fmt"
)

// ESError 定义了标准的 Elasticsearch 错误结构，包含错误类型、具体原因等详细信息。
type ESError struct {
	StatusCode int              // HTTP状态码
	Type       string           // 错误类型
	Reason     string           // 错误原因
	RootCause  []map[string]any // 根本原因
	RawBody    []byte           // 原始响应体
}

// Error 记录一条 Error 级别的日志信息。
func (e *ESError) Error() string {
	return fmt.Sprintf("ES错误 [%d]: %s - %s", e.StatusCode, e.Type, e.Reason)
}

// IsNotFound 判断是否为 404 错误
func (e *ESError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsConflict 判断是否为冲突错误
func (e *ESError) IsConflict() bool {
	return e.StatusCode == 409
}

// IsBadRequest 判断是否为bad请求错误
func (e *ESError) IsBadRequest() bool {
	return e.StatusCode == 400
}

// IsTimeout 判断是否为超时错误
func (e *ESError) IsTimeout() bool {
	return e.StatusCode == 408 || e.Type == "timeout_exception"
}

// ParseESError 解析 Elasticsearch 返回的原始响应体和 HTTP 状态码，并转换为结构化的 ESError 对象，方便程序进行错误类型的判断 (如 IsNotFound, IsConflict 等)。
func ParseESError(statusCode int, body []byte) *ESError {
	var errResp struct {
		Error struct {
			Type      string           `json:"type"`
			Reason    string           `json:"reason"`
			RootCause []map[string]any `json:"root_cause"`
		} `json:"error"`
		Status int `json:"status"`
	}
	err := json.Unmarshal(body, &errResp)
	if err != nil {
		return &ESError{
			StatusCode: statusCode,
			RawBody:    body,
		}
	}
	return &ESError{
		StatusCode: statusCode,
		Type:       errResp.Error.Type,
		Reason:     errResp.Error.Reason,
		RootCause:  errResp.Error.RootCause,
		RawBody:    body,
	}
}
