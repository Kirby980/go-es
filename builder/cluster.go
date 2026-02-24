package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ClusterBuilder 集群管理构建器
type ClusterBuilder struct {
	client ESClient
	DebugHelper
}

// NewClusterBuilder 创建集群构建器
func NewClusterBuilder(c ESClient) *ClusterBuilder {
	return &ClusterBuilder{
		client:      c,
		DebugHelper: DebugHelper{logger: c.GetLogger()},
	}
}

// Debug 启用调试模式（链式调用）
func (b *ClusterBuilder) Debug() *ClusterBuilder {
	b.SetDebug(true)
	return b
}

// ========== 集群健康 ==========

// ClusterHealthResponse 集群健康响应
type ClusterHealthResponse struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"` // green, yellow, red
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

// Health 获取集群健康状态
func (b *ClusterBuilder) Health(ctx context.Context) (*ClusterHealthResponse, error) {
	path := "/_cluster/health"

	// 如果启用调试模式，打印请求信息
	if b.IsDebug() {
		b.PrintDebug("GET", path, nil)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.IsDebug() {
		b.PrintResponse(respBody)
	}

	var resp ClusterHealthResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// IndexHealth 获取索引健康状态
func (b *ClusterBuilder) IndexHealth(ctx context.Context, index string) (*ClusterHealthResponse, error) {
	path := fmt.Sprintf("/_cluster/health/%s", url.PathEscape(index))
	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ClusterHealthResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// ========== 集群状态 ==========

// ClusterStateResponse 集群状态响应
type ClusterStateResponse struct {
	ClusterName  string         `json:"cluster_name"`
	ClusterUUID  string         `json:"cluster_uuid"`
	Version      int            `json:"version"`
	StateUUID    string         `json:"state_uuid"`
	MasterNode   string         `json:"master_node"`
	Blocks       map[string]any `json:"blocks"`
	Nodes        map[string]any `json:"nodes"`
	Metadata     map[string]any `json:"metadata"`
	RoutingTable map[string]any `json:"routing_table"`
	RoutingNodes map[string]any `json:"routing_nodes"`
}

// State 获取集群状态
func (b *ClusterBuilder) State(ctx context.Context) (*ClusterStateResponse, error) {
	path := "/_cluster/state"
	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ClusterStateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// ========== 集群统计 ==========

// ClusterStatsResponse 集群统计响应
type ClusterStatsResponse struct {
	ClusterName string `json:"cluster_name"`
	ClusterUUID string `json:"cluster_uuid"`
	Timestamp   int64  `json:"timestamp"`
	Status      string `json:"status"`
	Indices     struct {
		Count      int            `json:"count"`
		Shards     map[string]any `json:"shards"`
		Docs       map[string]any `json:"docs"`
		Store      map[string]any `json:"store"`
		FieldData  map[string]any `json:"fielddata"`
		QueryCache map[string]any `json:"query_cache"`
		Completion map[string]any `json:"completion"`
		Segments   map[string]any `json:"segments"`
	} `json:"indices"`
	Nodes map[string]any `json:"nodes"`
}

// Stats 获取集群统计信息
func (b *ClusterBuilder) Stats(ctx context.Context) (*ClusterStatsResponse, error) {
	path := "/_cluster/stats"

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.PrintDebug("GET", path, nil)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.PrintResponse(respBody)
	}

	var resp ClusterStatsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// ========== 节点信息 ==========

// NodesInfoResponse 节点信息响应
type NodesInfoResponse struct {
	ClusterName string                    `json:"cluster_name"`
	Nodes       map[string]map[string]any `json:"nodes"`
}

// NodesInfo 获取节点信息
func (b *ClusterBuilder) NodesInfo(ctx context.Context) (*NodesInfoResponse, error) {
	path := "/_nodes"
	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp NodesInfoResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// ========== 节点统计 ==========

// NodesStatsResponse 节点统计响应
type NodesStatsResponse struct {
	ClusterName string                    `json:"cluster_name"`
	Nodes       map[string]map[string]any `json:"nodes"`
}

// NodesStats 获取节点统计
func (b *ClusterBuilder) NodesStats(ctx context.Context) (*NodesStatsResponse, error) {
	path := "/_nodes/stats"
	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp NodesStatsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// ========== 任务管理 ==========

// TasksResponse 任务响应
type TasksResponse struct {
	Nodes map[string]any `json:"nodes"`
	Tasks map[string]any `json:"tasks"`
}

// Tasks 获取正在运行的任务
func (b *ClusterBuilder) Tasks(ctx context.Context) (*TasksResponse, error) {
	path := "/_tasks"
	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp TasksResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// ========== 集群设置 ==========

// ClusterSettingsResponse 集群设置响应
type ClusterSettingsResponse struct {
	Persistent map[string]any `json:"persistent"`
	Transient  map[string]any `json:"transient"`
}

// GetSettings 获取集群设置
func (b *ClusterBuilder) GetSettings(ctx context.Context) (*ClusterSettingsResponse, error) {
	path := "/_cluster/settings"
	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ClusterSettingsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// UpdateSettings 更新集群设置
func (b *ClusterBuilder) UpdateSettings(ctx context.Context, persistent, transient map[string]any) error {
	path := "/_cluster/settings"
	body := make(map[string]any)
	if persistent != nil {
		body["persistent"] = persistent
	}
	if transient != nil {
		body["transient"] = transient
	}

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.PrintDebug("PUT", path, body)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.PrintResponse(respBody)
	}

	return nil
}

// ========== 分配解释 ==========

// AllocationExplainResponse 分配解释响应
type AllocationExplainResponse struct {
	Index                   string           `json:"index"`
	Shard                   int              `json:"shard"`
	Primary                 bool             `json:"primary"`
	CurrentState            string           `json:"current_state"`
	UnassignedInfo          map[string]any   `json:"unassigned_info,omitempty"`
	CanAllocate             string           `json:"can_allocate,omitempty"`
	AllocateExplanation     string           `json:"allocate_explanation,omitempty"`
	NodeAllocationDecisions []map[string]any `json:"node_allocation_decisions,omitempty"`
}

// AllocationExplain 解释分片分配
func (b *ClusterBuilder) AllocationExplain(ctx context.Context, index string, shard int, primary bool) (*AllocationExplainResponse, error) {
	path := "/_cluster/allocation/explain"
	body := map[string]any{
		"index":   index,
		"shard":   shard,
		"primary": primary,
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	var resp AllocationExplainResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// ========== 远程集群 ==========

// RemoteClustersResponse 远程集群响应
type RemoteClustersResponse map[string]map[string]any

// RemoteClusters 获取远程集群信息
func (b *ClusterBuilder) RemoteClusters(ctx context.Context) (RemoteClustersResponse, error) {
	path := "/_remote/info"
	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp RemoteClustersResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return resp, nil
}
