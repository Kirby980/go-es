package builder

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// ShardsInfo 分片信息（通用）
type ShardsInfo struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

// HitsTotal 命中总数（通用）
type HitsTotal struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

// HitItem 单个命中文档（通用）
type HitItem struct {
	Index     string              `json:"_index"`
	ID        string              `json:"_id"`
	Score     float64             `json:"_score"`
	Source    map[string]any      `json:"_source"`
	Sort      []any               `json:"sort,omitempty"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}

// scanHits 将 HitItem 切片的 Source 扫描到目标结构体切片中
func scanHits(hits []HitItem, dest any) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to slice")
	}
	sliceVal := destVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}
	sources := make([]map[string]any, 0, len(hits))
	for _, hit := range hits {
		if hit.Source == nil {
			continue
		}
		sources = append(sources, hit.Source)
	}
	jsonData, err := json.Marshal(sources)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, dest)
}
