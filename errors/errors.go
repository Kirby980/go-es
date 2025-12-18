package errors

import (
	"encoding/json"
	"fmt"
)

type ESError struct {
	StatusCode int                      // HTTP状态码
	Type       string                   // 错误类型
	Reason     string                   // 错误原因
	RootCause  []map[string]interface{} // 根本原因
	RawBody    []byte                   // 原始响应体
}

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

func ParseESError(statusCode int, body []byte) *ESError {
	var errResp struct {
		Error struct {
			Type      string                   `json:"type"`
			Reason    string                   `json:"reason"`
			RootCause []map[string]interface{} `json:"root_cause"`
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
