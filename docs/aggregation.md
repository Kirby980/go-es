# 聚合分析 (AggregationBuilder)

AggregationBuilder 提供了完整的 Elasticsearch 聚合分析功能。

## 指标聚合

### 统计聚合

```go
aggResp, err := builder.NewAggregationBuilder(esClient, "products").
    Avg("avg_price", "price").
    Sum("total_quantity", "quantity").
    Min("min_price", "price").
    Max("max_price", "price").
    Stats("price_stats", "price").
    Cardinality("unique_categories", "category").
    Do(ctx)

fmt.Println(aggResp.PrettyJSON())
```

### 百分位聚合

```go
aggResp, err := builder.NewAggregationBuilder(esClient, "products").
    Percentiles("price_percentiles", "price").
    Do(ctx)
```

## 桶聚合

### 分组聚合

```go
// 基础分组
termsResp, err := builder.NewAggregationBuilder(esClient, "products").
    Terms("by_category", "category", 10).
    Do(ctx)

// 带排序的分组
termsResp, err := builder.NewAggregationBuilder(esClient, "products").
    TermsWithOrder("top_categories", "category", 5, "_count", "desc").
    Do(ctx)
```

### 直方图

```go
// 数值直方图
histResp, err := builder.NewAggregationBuilder(esClient, "products").
    Histogram("price_distribution", "price", 100).
    Do(ctx)

// 日期直方图
dateHistResp, err := builder.NewAggregationBuilder(esClient, "orders").
    DateHistogram("orders_over_time", "created_at", "1d").
    Do(ctx)
```

### 范围聚合

```go
rangeResp, err := builder.NewAggregationBuilder(esClient, "products").
    Range("price_ranges", "price", []map[string]interface{}{
        {"key": "cheap", "to": 300},
        {"key": "medium", "from": 300, "to": 1000},
        {"key": "expensive", "from": 1000},
    }).
    Do(ctx)
```

### 日期范围聚合

```go
dateRangeResp, err := builder.NewAggregationBuilder(esClient, "orders").
    DateRange("date_ranges", "created_at", []map[string]interface{}{
        {"key": "last_week", "from": "now-7d/d", "to": "now"},
        {"key": "last_month", "from": "now-1M/M", "to": "now"},
    }).
    Do(ctx)
```

## 过滤器聚合

```go
// 单过滤器
aggResp, err := builder.NewAggregationBuilder(esClient, "products").
    Filter("expensive_products", map[string]interface{}{
        "range": map[string]interface{}{
            "price": map[string]interface{}{"gte": 1000},
        },
    }).
    Do(ctx)

// 多过滤器
aggResp, err := builder.NewAggregationBuilder(esClient, "products").
    Filters("price_categories", map[string]interface{}{
        "cheap": map[string]interface{}{
            "range": map[string]interface{}{"price": map[string]interface{}{"lt": 300}},
        },
        "expensive": map[string]interface{}{
            "range": map[string]interface{}{"price": map[string]interface{}{"gte": 1000}},
        },
    }).
    Do(ctx)
```

## 管道聚合

### 平均桶聚合

```go
aggResp, err := builder.NewAggregationBuilder(esClient, "sales").
    DateHistogram("sales_per_month", "date", "1M").
    AvgBucket("avg_monthly_sales", "sales_per_month>_count").
    Do(ctx)
```

### 总和桶聚合

```go
aggResp, err := builder.NewAggregationBuilder(esClient, "sales").
    DateHistogram("sales_per_month", "date", "1M").
    SumBucket("total_sales", "sales_per_month>total_amount").
    Do(ctx)
```

### 累计求和

```go
cumulativeResp, err := builder.NewAggregationBuilder(esClient, "sales").
    DateHistogram("daily_sales", "date", "1d").
    CumulativeSum("cumulative_sales", "daily_sales>total_amount").
    Do(ctx)
```

### 移动平均

```go
aggResp, err := builder.NewAggregationBuilder(esClient, "sales").
    DateHistogram("daily_sales", "date", "1d").
    MovingAvg("sales_moving_avg", "daily_sales>total_amount", "simple", 7).
    Do(ctx)
```

### 导数

```go
aggResp, err := builder.NewAggregationBuilder(esClient, "sales").
    DateHistogram("daily_sales", "date", "1d").
    Derivative("sales_derivative", "daily_sales>total_amount").
    Do(ctx)
```

## 地理聚合

```go
// 地理边界
geoBoundsResp, err := builder.NewAggregationBuilder(esClient, "stores").
    GeoBounds("store_bounds", "location").
    Do(ctx)

// 地理中心点
geoCentroidResp, err := builder.NewAggregationBuilder(esClient, "stores").
    GeoCentroid("store_center", "location").
    Do(ctx)

// 地理距离聚合
geoDistResp, err := builder.NewAggregationBuilder(esClient, "stores").
    GeoDistance("distance_ranges", "location", 37.7749, -122.4194, []string{"0-100km", "100-300km", "300-*"}).
    Do(ctx)
```

## 支持的功能

- ✅ 指标聚合 (Avg, Sum, Min, Max, Count, Stats, Cardinality, Percentiles)
- ✅ 桶聚合 (Terms, Histogram, DateHistogram, Range, DateRange)
- ✅ 过滤器聚合 (Filter, Filters)
- ✅ 管道聚合 (AvgBucket, SumBucket, MovingAvg, Derivative, CumulativeSum)
- ✅ 地理聚合 (GeoBounds, GeoCentroid, GeoDistance)
