package builder

import (
	"encoding/json"

	"github.com/Kirby980/go-es/logger"
)

// debugHelper 调试辅助结构，嵌入各 Builder 中（包内私有）
type debugHelper struct {
	debug           bool
	debugPersistent bool
	logger          logger.Logger
}

func (d *debugHelper) isDebug() bool {
	return d.debug
}

func (d *debugHelper) setDebug(enabled bool) {
	if !enabled {
		d.debug = false
		d.debugPersistent = false
		return
	}
	d.debug = true
	d.debugPersistent = false
}

func (d *debugHelper) setDebugPersistent(enabled bool) {
	if !enabled {
		d.debugPersistent = false
		d.debug = false
		return
	}
	d.debugPersistent = true
	d.debug = true
}

// DebugPersistent 开启或关闭持久调试模式。开启时 (true)，同一构建器的所有后续请求都将被打印；关闭时 (false) 则停止打印。
func (d *debugHelper) DebugPersistent(enable bool) {
	d.setDebugPersistent(enable)
}

func (d *debugHelper) autoResetDebug() {
	if !d.debugPersistent {
		d.debug = false
	}
}

// log 返回 logger，若未初始化则返回 NopLogger，避免 nil panic
func (d *debugHelper) log() logger.Logger {
	if d.logger == nil {
		return logger.NopLogger{}
	}
	return d.logger
}

// printDebug 打印请求信息（仅在 debug 模式下调用）
func (d *debugHelper) printDebug(method, path string, body any) {
	log := d.log()
	if body != nil {
		log.Info("[ES Debug] 请求", "method", method, "path", path, "body", body)
	} else {
		log.Info("[ES Debug] 请求", "method", method, "path", path)
	}
}

// printResponse 打印响应体（仅在 debug 模式下调用）
func (d *debugHelper) printResponse(respBody []byte) {
	var bodyObj any
	if err := json.Unmarshal(respBody, &bodyObj); err == nil {
		d.log().Info("[ES Debug] 响应", "body", bodyObj)
	}
}

func (d *debugHelper) printResponseObj(obj any) {
	d.log().Info("[ES Debug] 响应", "body", obj)
}
