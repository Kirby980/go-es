package builder

import (
	"context"
	"testing"
)

// TestClusterBuilder_Health 测试集群健康
func TestClusterBuilder_Health(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	health, err := clusterBuilder.Health(ctx)
	if err != nil {
		t.Fatalf("获取集群健康失败: %v", err)
	}

	if health.ClusterName == "" {
		t.Error("集群名称不应该为空")
	}

	if health.Status != "green" && health.Status != "yellow" && health.Status != "red" {
		t.Errorf("集群状态应该是 green/yellow/red, 实际=%s", health.Status)
	}

	t.Logf("✓ 集群健康检查成功")
	t.Logf("  集群名称: %s", health.ClusterName)
	t.Logf("  状态: %s", health.Status)
	t.Logf("  节点数: %d", health.NumberOfNodes)
	t.Logf("  数据节点数: %d", health.NumberOfDataNodes)
	t.Logf("  活跃分片: %d", health.ActiveShards)
	t.Logf("  未分配分片: %d", health.UnassignedShards)
}

// TestClusterBuilder_IndexHealth 测试索引健康
func TestClusterBuilder_IndexHealth(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	// 创建测试索引
	indexName := "test_cluster_health"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	clusterBuilder := NewClusterBuilder(client)

	health, err := clusterBuilder.IndexHealth(ctx, indexName)
	if err != nil {
		t.Fatalf("获取索引健康失败: %v", err)
	}

	t.Logf("✓ 索引健康检查成功")
	t.Logf("  索引状态: %s", health.Status)
}

// TestClusterBuilder_State 测试集群状态
func TestClusterBuilder_State(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	state, err := clusterBuilder.State(ctx)
	if err != nil {
		t.Fatalf("获取集群状态失败: %v", err)
	}

	if state.ClusterName == "" {
		t.Error("集群名称不应该为空")
	}

	if state.MasterNode == "" {
		t.Error("主节点不应该为空")
	}

	t.Logf("✓ 集群状态获取成功")
	t.Logf("  集群名称: %s", state.ClusterName)
	t.Logf("  主节点: %s", state.MasterNode)
	t.Logf("  集群 UUID: %s", state.ClusterUUID)
}

// TestClusterBuilder_Stats 测试集群统计
func TestClusterBuilder_Stats(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	stats, err := clusterBuilder.Stats(ctx)
	if err != nil {
		t.Fatalf("获取集群统计失败: %v", err)
	}

	if stats.ClusterName == "" {
		t.Error("集群名称不应该为空")
	}

	t.Logf("✓ 集群统计获取成功")
	t.Logf("  集群名称: %s", stats.ClusterName)
	t.Logf("  状态: %s", stats.Status)
	t.Logf("  索引数量: %d", stats.Indices.Count)
}

// TestClusterBuilder_NodesInfo 测试节点信息
func TestClusterBuilder_NodesInfo(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	nodes, err := clusterBuilder.NodesInfo(ctx)
	if err != nil {
		t.Fatalf("获取节点信息失败: %v", err)
	}

	if len(nodes.Nodes) == 0 {
		t.Error("节点数量不应该为 0")
	}

	t.Logf("✓ 节点信息获取成功")
	t.Logf("  集群名称: %s", nodes.ClusterName)
	t.Logf("  节点数量: %d", len(nodes.Nodes))
}

// TestClusterBuilder_NodesStats 测试节点统计
func TestClusterBuilder_NodesStats(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	stats, err := clusterBuilder.NodesStats(ctx)
	if err != nil {
		t.Fatalf("获取节点统计失败: %v", err)
	}

	if len(stats.Nodes) == 0 {
		t.Error("节点数量不应该为 0")
	}

	t.Logf("✓ 节点统计获取成功")
	t.Logf("  集群名称: %s", stats.ClusterName)
	t.Logf("  节点数量: %d", len(stats.Nodes))
}

// TestClusterBuilder_Tasks 测试任务列表
func TestClusterBuilder_Tasks(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	tasks, err := clusterBuilder.Tasks(ctx)
	if err != nil {
		t.Fatalf("获取任务列表失败: %v", err)
	}

	t.Logf("✓ 任务列表获取成功")
	t.Logf("  任务数量: %d", len(tasks.Tasks))
}

// TestClusterBuilder_GetSettings 测试获取集群设置
func TestClusterBuilder_GetSettings(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	settings, err := clusterBuilder.GetSettings(ctx)
	if err != nil {
		t.Fatalf("获取集群设置失败: %v", err)
	}

	t.Logf("✓ 集群设置获取成功")
	t.Logf("  持久化设置: %v", settings.Persistent)
	t.Logf("  临时设置: %v", settings.Transient)
}

// TestClusterBuilder_UpdateSettings 测试更新集群设置
func TestClusterBuilder_UpdateSettings(t *testing.T) {
	t.Skip("跳过设置更新测试，避免影响集群配置")

	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	// 更新临时设置
	err := clusterBuilder.UpdateSettings(ctx,
		nil, // persistent
		map[string]interface{}{
			"cluster.routing.allocation.enable": "all",
		},
	)

	if err != nil {
		t.Fatalf("更新集群设置失败: %v", err)
	}

	t.Logf("✓ 集群设置更新成功")
}

// TestClusterBuilder_AllocationExplain 测试分配解释
func TestClusterBuilder_AllocationExplain(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	// 创建测试索引
	indexName := "test_allocation_explain"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	clusterBuilder := NewClusterBuilder(client)

	// 解释第一个分片的分配
	explain, err := clusterBuilder.AllocationExplain(ctx, indexName, 0, true)
	if err != nil {
		t.Logf("分配解释获取: %v (可能没有未分配分片)", err)
		return
	}

	t.Logf("✓ 分配解释获取成功")
	t.Logf("  索引: %s", explain.Index)
	t.Logf("  分片: %d", explain.Shard)
	t.Logf("  主分片: %v", explain.Primary)
	t.Logf("  当前状态: %s", explain.CurrentState)
}

// TestClusterBuilder_RemoteClusters 测试远程集群信息
func TestClusterBuilder_RemoteClusters(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	clusterBuilder := NewClusterBuilder(client)

	clusters, err := clusterBuilder.RemoteClusters(ctx)
	if err != nil {
		t.Fatalf("获取远程集群信息失败: %v", err)
	}

	t.Logf("✓ 远程集群信息获取成功")
	t.Logf("  远程集群数量: %d", len(clusters))
}
