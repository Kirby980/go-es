# 索引管理 (IndexBuilder)

IndexBuilder 提供了完整的 Elasticsearch 索引管理功能，包括创建、更新、查询和删除索引。

## AutoMigrate（推荐）

类似 GORM 的自动迁移功能，通过结构体标签自动创建或更新索引映射。

### 基本用法

```go
// 定义结构体，使用 es 标签定义字段映射
type Article struct {
    Title     string    `es:"type:text;analyzer:ik_smart"`
    Content   string    `es:"type:text;analyzer:ik_max_word"`
    Author    string    `es:"type:keyword"`
    Tags      []string  `es:"type:keyword"`
    Views     int       `es:"type:integer"`
    Price     float64   `es:"type:float"`
    Published bool      `es:"type:boolean"`
    CreatedAt time.Time `es:"type:date;format:yyyy-MM-dd HH:mm:ss"`
    Location  string    `es:"type:geo_point"`
}

// 自动迁移（创建或更新索引）
import "github.com/Kirby980/go-es/sugar"

err := sugar.New(esClient).AutoMigrate(&Article{})
```

### 自定义索引名

默认索引名为结构体名称的 snake_case 复数形式（如 `Article` -> `articles`）。

可以实现 `IndexName` 接口自定义索引名：

```go
type Article struct {
    // ...
}

// 实现 IndexName 接口
func (a *Article) IndexName() string {
    return "my_articles"
}

// 迁移时会使用 "my_articles" 作为索引名
import "github.com/Kirby980/go-es/sugar"

err := sugar.New(esClient).AutoMigrate(&Article{})
```

### es 标签语法

标签格式：`es:"key1:value1;key2:value2"`

常用配置：
- `type:xxx` - 字段类型（必填）
- `analyzer:xxx` - 分析器
- `format:xxx` - 日期格式
- `-` - 忽略该字段

```go
type Product struct {
    Name        string  `es:"type:text;analyzer:ik_smart"`     // text + IK分词
    SKU         string  `es:"type:keyword"`                     // 精确匹配
    Price       float64 `es:"type:float"`                       // 浮点数
    Stock       int     `es:"type:integer"`                     // 整数
    Available   bool    `es:"type:boolean"`                     // 布尔
    CreatedAt   string  `es:"type:date;format:yyyy-MM-dd"`      // 日期
    Location    string  `es:"type:geo_point"`                   // 地理位置
    InternalID  string  `es:"-"`                                // 忽略
}
```

### 迁移多个模型

```go
import "github.com/Kirby980/go-es/sugar"

err := sugar.New(esClient).AutoMigrate(
    &Article{},
    &Product{},
    &User{},
)
```

### 注意事项

- ES 不支持删除或修改已有字段类型，AutoMigrate 只能添加新字段
- 如果索引已存在，会调用 PutMapping 添加新字段
- 如果索引不存在，会创建新索引

## 基础操作

### 创建索引

```go
// 基础创建
err := builder.NewIndexBuilder(esClient, "products").
    Shards(1).
    Replicas(0).
    RefreshInterval("1s").
    AddProperty("name", "text", builder.WithAnalyzer("ik_smart")).
    AddProperty("price", "float").
    AddProperty("category", "keyword").
    AddProperty("created_at", "date", builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
    AddAlias("products-alias", nil).
    Create(ctx)
```

### 更新索引设置

```go
err := builder.NewIndexBuilder(esClient, "products").
    Replicas(2).
    RefreshInterval("30s").
    UpdateSettings(ctx)
```

### 更新索引映射

```go
// 添加新字段
err := builder.NewIndexBuilder(esClient, "products").
    AddProperty("description", "text", builder.WithAnalyzer("ik_max_word")).
    AddProperty("stock", "integer").
    PutMapping(ctx)
```

### 检查索引是否存在

```go
exists, err := builder.NewIndexBuilder(esClient, "products").Exists(ctx)
```

### 获取索引信息

```go
info, err := builder.NewIndexBuilder(esClient, "products").Get(ctx)
fmt.Println(info.PrettyJSON())
```

### 删除索引

```go
err := builder.NewIndexBuilder(esClient, "products").Delete(ctx)
```

## 自定义 Header

可以通过 `.Header(key, value)` 方法为单次请求添加自定义 HTTP Header：

```go
err := builder.NewIndexBuilder(esClient, "products").
    Header("X-Custom-Source", "my-app").
    Exists(ctx)
```

## 自定义分析器

### 方式1：简化版（基于 tokenizer 快速创建）

```go
import esconst "github.com/Kirby980/go-es/const"

err := builder.NewIndexBuilder(esClient, "articles").
    Shards(1).
    Replicas(1).
    // 基于 tokenizer 快速创建（不忽略大小写的 ik_smart）
    AddCustomAnalyzer("ik_case_sensitive", esconst.TokenizerIKSmart).
    AddProperty("title", "text", builder.WithAnalyzer("ik_case_sensitive")).
    Create(ctx)
```

### 方式2：完整版（使用 Option 模式自定义配置）

```go
import esconst "github.com/Kirby980/go-es/const"

err := builder.NewIndexBuilder(esClient, "articles").
    AddAnalyzer("html_ik_analyzer",
        builder.WithAnalyzerType(esconst.AnalyzerTypeCustom),
        builder.WithCharFilters(esconst.CharFilterHTMLStrip),  // 去除HTML标签
        builder.WithTokenizer(esconst.TokenizerIKSmart),       // IK分词
        builder.WithTokenFilters(esconst.TokenFilterLowercase), // 转小写
    ).
    AddProperty("content", "text", builder.WithAnalyzer("html_ik_analyzer")).
    Create(ctx)
```

### 常用分析器示例

```go
import esconst "github.com/Kirby980/go-es/const"

err := builder.NewIndexBuilder(esClient, "posts").
    // IK 智能分词（保持大小写）
    AddCustomAnalyzer("ik_case_sensitive", esconst.TokenizerIKSmart).

    // IK 最大词数分词（转小写）
    AddCustomAnalyzer("ik_lowercase", esconst.TokenizerIKMaxWord,
        esconst.TokenFilterLowercase).

    // 标准分词器 + 词干提取 + 停用词过滤
    AddAnalyzer("english_stemmed",
        builder.WithAnalyzerType(esconst.AnalyzerTypeCustom),
        builder.WithTokenizer(esconst.TokenizerStandard),
        builder.WithTokenFilters(
            esconst.TokenFilterLowercase,
            esconst.TokenFilterStop,
            esconst.TokenFilterPorterStem,
        ),
    ).
    Create(ctx)
```

## 分析器常量说明

所有可用的常量定义在 `const` 包中：

### 分词器常量 (Tokenizers)

```go
esconst.TokenizerIKSmart      // IK智能分词（粗粒度）
esconst.TokenizerIKMaxWord    // IK最大词数分词（细粒度）
esconst.TokenizerStandard     // 标准分词器
esconst.TokenizerWhitespace   // 空格分词器
esconst.TokenizerKeyword      // 关键词分词器（不分词）
// 更多见 const 包
```

### 字符过滤器常量 (Char Filters)

```go
esconst.CharFilterHTMLStrip   // 去除HTML标签
esconst.CharFilterMapping     // 字符映射替换
esconst.CharFilterPattern     // 正则替换
```

### Token过滤器常量 (Token Filters)

```go
esconst.TokenFilterLowercase  // 转小写
esconst.TokenFilterUppercase  // 转大写
esconst.TokenFilterStop       // 去除停用词
esconst.TokenFilterStemmer    // 词干提取
esconst.TokenFilterSynonym    // 同义词替换
esconst.TokenFilterUnique     // 去重
// 更多见 const 包
```

### 分析器类型常量 (Analyzer Types)

```go
esconst.AnalyzerTypeCustom    // 自定义分析器
esconst.AnalyzerTypeStandard  // 标准分析器
esconst.AnalyzerIKSmart       // IK智能分词器（内置）
esconst.AnalyzerIKMaxWord     // IK最大词数分词器（内置）
```

## 字段类型常量说明

所有可用的字段类型常量定义在 `const` 包中：

### 字符串类型

```go
esconst.FieldTypeText      // 全文搜索字段（会分词）
esconst.FieldTypeKeyword   // 精确匹配字段（不分词）
```

### 数值类型

```go
esconst.FieldTypeInt       // 32位整数
esconst.FieldTypeLong      // 64位整数
esconst.FieldTypeFloat     // 32位浮点数
esconst.FieldTypeDouble    // 64位双精度浮点数
esconst.FieldTypeByte      // 8位整数
esconst.FieldTypeShort     // 16位整数
```

### 日期/布尔类型

```go
esconst.FieldTypeDate      // 日期类型
esconst.FieldTypeBoolean   // 布尔类型
```

### 地理位置类型

```go
esconst.FieldTypeGeoPoint  // 地理位置点
esconst.FieldTypeGeoShape  // 地理形状
```

### 复杂类型

```go
esconst.FieldTypeObject    // 对象类型
esconst.FieldTypeNested    // 嵌套类型
```

### 其他类型

```go
esconst.FieldTypeIP              // IP地址
esconst.FieldTypeBinary          // 二进制
esconst.FieldTypeCompletion      // 自动补全
esconst.FieldTypeDenseVector     // 密集向量（向量搜索）
// 更多见 const 包
```

### 使用示例（使用常量避免拼写错误）

```go
import esconst "github.com/Kirby980/go-es/const"

err := builder.NewIndexBuilder(esClient, "products").
    // 使用字段类型常量，IDE 会自动提示，避免拼写错误
    AddProperty("name", esconst.FieldTypeText, builder.WithAnalyzer(esconst.AnalyzerIKSmart)).
    AddProperty("sku", esconst.FieldTypeKeyword).
    AddProperty("price", esconst.FieldTypeFloat).
    AddProperty("quantity", esconst.FieldTypeInt).
    AddProperty("available", esconst.FieldTypeBoolean).
    AddProperty("created_at", esconst.FieldTypeDate, builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
    AddProperty("location", esconst.FieldTypeGeoPoint).
    Create(ctx)
```

## 支持的功能

- ✅ AutoMigrate（自动迁移）
- ✅ 创建索引 (Create, Shards, Replicas, RefreshInterval)
- ✅ 字段映射 (AddProperty, WithAnalyzer, WithFormat)
- ✅ 别名管理 (AddAlias)
- ✅ 更新索引设置 (UpdateSettings)
- ✅ 更新索引映射 (PutMapping)
- ✅ 检查存在 (Exists)
- ✅ 获取索引信息 (Get)
- ✅ 删除索引 (Delete)
