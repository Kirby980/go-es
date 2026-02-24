package builder

import (
	"encoding/json"

	"github.com/Kirby980/go-es/logger"
)

// Debugger 调试接口
type Debugger interface {
	IsDebug() bool
	SetDebug(bool)
	PrintDebug(method, path string, body any)
	PrintResponse(respBody []byte)
}

// DebugHelper 调试辅助结构，嵌入各 Builder 中
type DebugHelper struct {
	debug  bool
	logger logger.Logger
}

func (d *DebugHelper) IsDebug() bool {
	return d.debug
}

func (d *DebugHelper) SetDebug(enabled bool) {
	d.debug = enabled
}

// log 返回 logger，若未初始化则返回 NopLogger，避免 nil panic
func (d *DebugHelper) log() logger.Logger {
	if d.logger == nil {
		return logger.NopLogger{}
	}
	return d.logger
}

// PrintDebug 打印请求信息（仅在 debug 模式下调用）
// body 直接作为结构化字段传入 logger，zap 会将其序列化为嵌套 JSON 对象
func (d *DebugHelper) PrintDebug(method, path string, body any) {
	log := d.log()
	if body != nil {
		log.Info("[ES Debug] 请求", "method", method, "path", path, "body", body)
	} else {
		log.Info("[ES Debug] 请求", "method", method, "path", path)
	}
}

// PrintResponse 打印响应体（仅在 debug 模式下调用）
// 将原始 JSON bytes 反序列化为 any，让 logger 将其序列化为嵌套 JSON 对象
func (d *DebugHelper) PrintResponse(respBody []byte) {
	var bodyObj any
	if err := json.Unmarshal(respBody, &bodyObj); err == nil {
		d.log().Info("[ES Debug] 响应", "body", bodyObj)
	}
}
