package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AggregationBuilder 聚合构建器
type AggregationBuilder struct {
	client ESClient
	index  string
	query  map[string]any
	aggs   map[string]any
	size   int
	debugHelper
	baseBuilder
}

// NewAggregationBuilder 创建聚合构建器
func NewAggregationBuilder(c ESClient, index string) *AggregationBuilder {
	b := &AggregationBuilder{
		client:      c,
		index:       index,
		aggs:        make(map[string]any),
		size:        0, // 聚合时默认不返回文档
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header (链式调用)
func (b *AggregationBuilder) Header(key, value string) *AggregationBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Query 设置查询条件
func (b *AggregationBuilder) Query(query map[string]any) *AggregationBuilder {
	b.query = query
	return b
}

// Size 设置返回文档数量
func (b *AggregationBuilder) Size(size int) *AggregationBuilder {
	b.size = size
	return b
}

// ========== 指标聚合 ==========

// Avg 平均值聚合
func (b *AggregationBuilder) Avg(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"avg": map[string]any{
			"field": field,
		},
	}
	return b
}

// Sum 求和聚合
func (b *AggregationBuilder) Sum(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"sum": map[string]any{
			"field": field,
		},
	}
	return b
}

// Min 最小值聚合
func (b *AggregationBuilder) Min(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"min": map[string]any{
			"field": field,
		},
	}
	return b
}

// Max 最大值聚合
func (b *AggregationBuilder) Max(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"max": map[string]any{
			"field": field,
		},
	}
	return b
}

// Count 计数聚合
func (b *AggregationBuilder) Count(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"value_count": map[string]any{
			"field": field,
		},
	}
	return b
}

// Stats 统计聚合（count, min, max, avg, sum）
func (b *AggregationBuilder) Stats(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"stats": map[string]any{
			"field": field,
		},
	}
	return b
}

// ExtendedStats 扩展统计聚合
func (b *AggregationBuilder) ExtendedStats(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"extended_stats": map[string]any{
			"field": field,
		},
	}
	return b
}

// Cardinality 基数聚合（唯一值数量）
func (b *AggregationBuilder) Cardinality(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"cardinality": map[string]any{
			"field": field,
		},
	}
	return b
}

// Percentiles 百分位聚合
func (b *AggregationBuilder) Percentiles(name, field string, percents ...float64) *AggregationBuilder {
	agg := map[string]any{
		"field": field,
	}
	if len(percents) > 0 {
		agg["percents"] = percents
	}
	b.aggs[name] = map[string]any{
		"percentiles": agg,
	}
	return b
}

// ========== 桶聚合 ==========

// Terms 词条聚合（分组）
func (b *AggregationBuilder) Terms(name, field string, size int) *AggregationBuilder {
	termsAgg := map[string]any{
		"field": field,
	}
	if size > 0 {
		termsAgg["size"] = size
	}
	b.aggs[name] = map[string]any{
		"terms": termsAgg,
	}
	return b
}

// TermsWithOrder 带排序的词条聚合
func (b *AggregationBuilder) TermsWithOrder(name, field string, size int, orderBy string, order string) *AggregationBuilder {
	termsAgg := map[string]any{
		"field": field,
		"order": map[string]any{
			orderBy: order,
		},
	}
	if size > 0 {
		termsAgg["size"] = size
	}
	b.aggs[name] = map[string]any{
		"terms": termsAgg,
	}
	return b
}

// Histogram 直方图聚合
func (b *AggregationBuilder) Histogram(name, field string, interval float64) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"histogram": map[string]any{
			"field":    field,
			"interval": interval,
		},
	}
	return b
}

// DateHistogram 日期直方图聚合
func (b *AggregationBuilder) DateHistogram(name, field, interval string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"date_histogram": map[string]any{
			"field":             field,
			"calendar_interval": interval, // 1d, 1w, 1M, 1y 等
		},
	}
	return b
}

// DateHistogramFixed 固定间隔日期直方图
func (b *AggregationBuilder) DateHistogramFixed(name, field, interval string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"date_histogram": map[string]any{
			"field":          field,
			"fixed_interval": interval, // 30s, 1m, 1h 等
		},
	}
	return b
}

// Range 范围聚合
func (b *AggregationBuilder) Range(name, field string, ranges []map[string]any) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"range": map[string]any{
			"field":  field,
			"ranges": ranges,
		},
	}
	return b
}

// DateRange 日期范围聚合
func (b *AggregationBuilder) DateRange(name, field string, ranges []map[string]any) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"date_range": map[string]any{
			"field":  field,
			"ranges": ranges,
		},
	}
	return b
}

// Filter 过滤器聚合
func (b *AggregationBuilder) Filter(name string, filter map[string]any) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"filter": filter,
	}
	return b
}

// Filters 多过滤器聚合
func (b *AggregationBuilder) Filters(name string, filters map[string]any) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"filters": map[string]any{
			"filters": filters,
		},
	}
	return b
}

// Missing 缺失值聚合
func (b *AggregationBuilder) Missing(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"missing": map[string]any{
			"field": field,
		},
	}
	return b
}

// ========== 嵌套聚合 ==========

// SubAgg 添加子聚合
func (b *AggregationBuilder) SubAgg(parentName string, subAgg map[string]any) *AggregationBuilder {
	if parent, ok := b.aggs[parentName]; ok {
		if parentMap, ok := parent.(map[string]any); ok {
			parentMap["aggs"] = subAgg
		}
	}
	return b
}

// ========== 管道聚合 ==========

// AvgBucket 平均桶聚合
func (b *AggregationBuilder) AvgBucket(name, bucketsPath string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"avg_bucket": map[string]any{
			"buckets_path": bucketsPath,
		},
	}
	return b
}

// SumBucket 求和桶聚合
func (b *AggregationBuilder) SumBucket(name, bucketsPath string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"sum_bucket": map[string]any{
			"buckets_path": bucketsPath,
		},
	}
	return b
}

// MaxBucket 最大桶聚合
func (b *AggregationBuilder) MaxBucket(name, bucketsPath string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"max_bucket": map[string]any{
			"buckets_path": bucketsPath,
		},
	}
	return b
}

// MinBucket 最小桶聚合
func (b *AggregationBuilder) MinBucket(name, bucketsPath string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"min_bucket": map[string]any{
			"buckets_path": bucketsPath,
		},
	}
	return b
}

// MovingAvg 移动平均聚合
func (b *AggregationBuilder) MovingAvg(name, bucketsPath string, window int) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"moving_avg": map[string]any{
			"buckets_path": bucketsPath,
			"window":       window,
		},
	}
	return b
}

// Derivative 导数聚合
func (b *AggregationBuilder) Derivative(name, bucketsPath string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"derivative": map[string]any{
			"buckets_path": bucketsPath,
		},
	}
	return b
}

// CumulativeSum 累计求和聚合
func (b *AggregationBuilder) CumulativeSum(name, bucketsPath string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"cumulative_sum": map[string]any{
			"buckets_path": bucketsPath,
		},
	}
	return b
}

// ========== 地理聚合 ==========

// GeoBounds 地理边界聚合
func (b *AggregationBuilder) GeoBounds(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"geo_bounds": map[string]any{
			"field": field,
		},
	}
	return b
}

// GeoCentroid 地理中心点聚合
func (b *AggregationBuilder) GeoCentroid(name, field string) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"geo_centroid": map[string]any{
			"field": field,
		},
	}
	return b
}

// GeoDistance 地理距离聚合
func (b *AggregationBuilder) GeoDistance(name, field string, origin map[string]float64, ranges []map[string]any) *AggregationBuilder {
	b.aggs[name] = map[string]any{
		"geo_distance": map[string]any{
			"field":  field,
			"origin": origin,
			"ranges": ranges,
		},
	}
	return b
}

// ========== 响应结构 ==========

// AggregationResponse 聚合响应
type AggregationResponse struct {
	Took         int            `json:"took"`
	TimedOut     bool           `json:"timed_out"`
	Shards       map[string]any `json:"_shards"`
	Hits         map[string]any `json:"hits"`
	Aggregations map[string]any `json:"aggregations"`
}

// Build 构建聚合请求
func (b *AggregationBuilder) Build() map[string]any {
	body := map[string]any{
		"size": b.size,
		"aggs": b.aggs,
	}

	if b.query != nil {
		body["query"] = b.query
	}

	return body
}

// Debug 启用调试模式（链式调用）
func (b *AggregationBuilder) Debug() *AggregationBuilder {
	b.setDebug(true)
	return b
}

// Do 执行聚合
func (b *AggregationBuilder) Do(ctx context.Context) (*AggregationResponse, error) {
	path := fmt.Sprintf("/%s/_search", url.PathEscape(b.index))
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp AggregationResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
