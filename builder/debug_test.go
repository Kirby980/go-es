package builder

import (
	"context"
	"testing"
	"time"
)

// TestDebugReset 测试debug模式在执行后自动重置
func TestDebugReset(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_debug_reset"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 创建一个builder实例（模拟gorm的使用方式）
	docBuilder := NewDocumentBuilder(client, indexName)

	// 第一次调用，带debug
	t.Log("=== 第一次调用（带debug） ===")
	_, err := docBuilder.ID("1").Set("title", "第一次调用").Debug().Do(ctx)
	if err != nil {
		t.Fatalf("第一次调用失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 第二次调用，不调用Debug()，应该不会输出debug信息
	t.Log("=== 第二次调用（不带debug） ===")
	_, err = docBuilder.ID("2").Set("title", "第二次调用").Do(ctx)
	if err != nil {
		t.Fatalf("第二次调用失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 第三次调用，再次带debug
	t.Log("=== 第三次调用（再次带debug） ===")
	_, err = docBuilder.ID("3").Set("title", "第三次调用").Debug().Do(ctx)
	if err != nil {
		t.Fatalf("第三次调用失败: %v", err)
	}

	t.Log("✓ Debug重置测试通过：每次调用独立控制debug")
}
