package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"
)

// 测试不同连接池配置的性能
// go test -bench=. -cpuprofile=cpu.prof ./benchmark/
// go prof -http=:8080 cpu.prof
// 1. 默认配置（小连接池）
func BenchmarkSearch_SmallPool(b *testing.B) {
	esClient, _ := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
		config.WithTimeout(10*time.Second),
		// 默认配置：MaxIdleConns=100, MaxIdleConnsPerHost=10
	)
	defer esClient.Close()

	ctx := context.Background()

	b.ResetTimer() // 重置计时器（不计入初始化时间）
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			builder.NewSearchBuilder(esClient, "products").
				Match("name", "test").
				Do(ctx)
		}
	})
}

// 2. 中等连接池
func BenchmarkSearch_MediumPool(b *testing.B) {
	esClient, _ := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
		config.WithTimeout(10*time.Second),
		config.WithMaxIdleConnsPerHost(50),
		config.WithMaxIdConns(200),
		config.WithMaxConnsPerHost(0), // 高并发需要大连接池
	)
	defer esClient.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			builder.NewSearchBuilder(esClient, "products").
				Match("name", "test").
				Do(ctx)
		}
	})
}

// 3. 大连接池
func BenchmarkSearch_LargePool(b *testing.B) {
	esClient, _ := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
		config.WithTimeout(10*time.Second),
		config.WithMaxIdleConnsPerHost(100),
		config.WithMaxIdConns(500),
		config.WithMaxConnsPerHost(0), // 高并发需要大连接池
	)
	defer esClient.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			builder.NewSearchBuilder(esClient, "products").
				Match("name", "test").
				Do(ctx)
		}
	})
}

// 4. 不同并发级别测试
func BenchmarkSearch_Concurrency10(b *testing.B) {
	esClient, _ := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
		config.WithMaxIdleConnsPerHost(50),
		config.WithMaxIdConns(200),
		config.WithMaxConnsPerHost(0), // 高并发需要大连接池
	)
	defer esClient.Close()

	ctx := context.Background()

	b.SetParallelism(10) // 10 个并发
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			builder.NewSearchBuilder(esClient, "products").
				Match("name", "test").
				Do(ctx)
		}
	})
}

func BenchmarkSearch_Concurrency100(b *testing.B) {
	esClient, _ := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
		config.WithMaxIdleConnsPerHost(50),
		config.WithMaxIdConns(200),
		config.WithMaxConnsPerHost(0), // 高并发需要大连接池
	)
	defer esClient.Close()

	ctx := context.Background()

	b.SetParallelism(100) // 100 个并发
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			builder.NewSearchBuilder(esClient, "products").
				Match("name", "test").
				Do(ctx)
		}
	})
}

func BenchmarkSearch_Concurrency1000(b *testing.B) {
	esClient, _ := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
		config.WithMaxIdleConnsPerHost(50),
		config.WithMaxIdConns(200),
		config.WithMaxConnsPerHost(0), // 高并发需要大连接池
	)
	defer esClient.Close()

	ctx := context.Background()

	b.SetParallelism(1000) // 1000 个并发
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			builder.NewSearchBuilder(esClient, "products").
				Match("name", "test").
				Do(ctx)
		}
	})
}
