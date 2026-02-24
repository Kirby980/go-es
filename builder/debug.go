package builder

import (
	"encoding/json"
	"fmt"
)

// Debugger 调试接口
type Debugger interface {
	IsDebug() bool
	SetDebug(bool)
	PrintDebug(method, path string, body any)
	PrintResponse(respBody []byte)
}

// DebugHelper 调试辅助结构
type DebugHelper struct {
	debug bool
}

func (d *DebugHelper) IsDebug() bool {
	return d.debug
}

func (d *DebugHelper) SetDebug(enabled bool) {
	d.debug = enabled
}

func (d *DebugHelper) PrintDebug(method, path string, body any) {
	fmt.Printf("\n[ES Debug] %s %s\n", method, path)
	if body != nil {
		data, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal body: %v\n", err)
		} else {
			fmt.Printf("Request Body:\n%s\n", string(data))
		}
	}
}

func (d *DebugHelper) PrintResponse(respBody []byte) {
	var pretty any
	if err := json.Unmarshal(respBody, &pretty); err == nil {
		data, err := json.MarshalIndent(pretty, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal response: %v\n", err)
		} else {
			fmt.Printf("Response:\n%s\n\n", string(data))
		}
	}
}
