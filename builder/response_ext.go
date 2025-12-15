package builder

import "encoding/json"

// ==================== Document Responses ====================

// JSON 返回 JSON 格式字符串（紧凑）
func (r *DocumentResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *DocumentResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *DocumentResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *GetResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *GetResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *GetResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *MGetResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *MGetResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *MGetResponse) String() string {
	return r.PrettyJSON()
}

// ==================== Bulk Response ====================

// JSON 返回 JSON 格式字符串（紧凑）
func (r *BulkResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *BulkResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *BulkResponse) String() string {
	return r.PrettyJSON()
}

// ==================== Search Responses ====================

// JSON 返回 JSON 格式字符串（紧凑）
func (r *SearchResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *SearchResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *SearchResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *CountResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *CountResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *CountResponse) String() string {
	return r.PrettyJSON()
}

// ==================== Query Responses ====================

// JSON 返回 JSON 格式字符串（紧凑）
func (r *DeleteByQueryResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *DeleteByQueryResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *DeleteByQueryResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *UpdateByQueryResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *UpdateByQueryResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *UpdateByQueryResponse) String() string {
	return r.PrettyJSON()
}

// ==================== Scroll Response ====================

// JSON 返回 JSON 格式字符串（紧凑）
func (r *ScrollResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *ScrollResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *ScrollResponse) String() string {
	return r.PrettyJSON()
}

// ==================== Aggregation Response ====================

// JSON 返回 JSON 格式字符串
func (r *AggregationResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *AggregationResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *AggregationResponse) String() string {
	return r.PrettyJSON()
}

// ==================== Index Response ====================

// JSON 返回 JSON 格式字符串（紧凑）
func (r *IndexInfo) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *IndexInfo) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *IndexInfo) String() string {
	return r.PrettyJSON()
}

// ==================== Cluster Responses ====================

// JSON 返回 JSON 格式字符串（紧凑）
func (r *ClusterHealthResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *ClusterHealthResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *ClusterHealthResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *ClusterStateResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *ClusterStateResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *ClusterStateResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *ClusterStatsResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *ClusterStatsResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *ClusterStatsResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *NodesInfoResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *NodesInfoResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *NodesInfoResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *NodesStatsResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *NodesStatsResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *NodesStatsResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *TasksResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *TasksResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *TasksResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *ClusterSettingsResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *ClusterSettingsResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *ClusterSettingsResponse) String() string {
	return r.PrettyJSON()
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *AllocationExplainResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *AllocationExplainResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *AllocationExplainResponse) String() string {
	return r.PrettyJSON()
}
