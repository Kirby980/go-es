package constant

// ===== 字符串类型 (String Types) =====

const (
	// FieldTypeText 全文搜索字段（会分词）
	// 适合：文章内容、标题、描述等需要全文搜索的字段
	FieldTypeText = "text"

	// FieldTypeKeyword 精确匹配字段（不分词）
	// 适合：标签、分类、状态、ID、邮箱等需要精确匹配、聚合、排序的字段
	FieldTypeKeyword = "keyword"
)

// ===== 数值类型 (Numeric Types) =====

const (
	// 整数类型
	FieldTypeByte  = "byte"    // 8位有符号整数：-128 到 127
	FieldTypeShort = "short"   // 16位有符号整数：-32,768 到 32,767
	FieldTypeInt   = "integer" // 32位有符号整数：-2^31 到 2^31-1
	FieldTypeLong  = "long"    // 64位有符号整数：-2^63 到 2^63-1

	// 浮点类型
	FieldTypeFloat  = "float"  // 32位单精度浮点数
	FieldTypeDouble = "double" // 64位双精度浮点数

	// 高精度类型
	FieldTypeHalfFloat    = "half_float"    // 16位半精度浮点数
	FieldTypeScaledFloat  = "scaled_float"  // 用固定比例因子缩放的浮点数
	FieldTypeUnsignedLong = "unsigned_long" // 64位无符号整数：0 到 2^64-1
)

// ===== 日期类型 (Date Types) =====

const (
	// FieldTypeDate 日期类型
	// 支持格式：ISO 8601, epoch_millis, 自定义格式
	FieldTypeDate = "date"

	// FieldTypeDateNanos 纳秒精度日期
	FieldTypeDateNanos = "date_nanos"
)

// ===== 布尔类型 (Boolean Type) =====

const (
	// FieldTypeBoolean 布尔类型
	// 接受：true, false, "true", "false"
	FieldTypeBoolean = "boolean"
)

// ===== 二进制类型 (Binary Type) =====

const (
	// FieldTypeBinary 二进制数据（Base64 编码）
	FieldTypeBinary = "binary"
)

// ===== 范围类型 (Range Types) =====

const (
	FieldTypeIntegerRange = "integer_range" // 整数范围
	FieldTypeLongRange    = "long_range"    // 长整数范围
	FieldTypeFloatRange   = "float_range"   // 浮点数范围
	FieldTypeDoubleRange  = "double_range"  // 双精度浮点数范围
	FieldTypeDateRange    = "date_range"    // 日期范围
	FieldTypeIPRange      = "ip_range"      // IP 地址范围
)

// ===== 复杂类型 (Complex Types) =====

const (
	// FieldTypeObject 对象类型（嵌套字段）
	// JSON 对象会被展平为点分隔的字段
	FieldTypeObject = "object"

	// FieldTypeNested 嵌套类型
	// 保持数组中对象的独立性，适合需要独立查询数组元素的场景
	FieldTypeNested = "nested"

	// FieldTypeFlattened 扁平化对象
	// 将整个对象作为单个字段索引，适合大量动态字段的场景
	FieldTypeFlattened = "flattened"
)

// ===== 地理位置类型 (Geo Types) =====

const (
	// FieldTypeGeoPoint 地理位置点（经纬度）
	// 格式：{"lat": 40.7128, "lon": -74.0060}
	FieldTypeGeoPoint = "geo_point"

	// FieldTypeGeoShape 地理形状（多边形、线等）
	FieldTypeGeoShape = "geo_shape"
)

// ===== IP 类型 (IP Type) =====

const (
	// FieldTypeIP IP 地址类型
	// 支持 IPv4 和 IPv6
	FieldTypeIP = "ip"
)

// ===== 自动补全类型 (Completion Type) =====

const (
	// FieldTypeCompletion 自动补全字段
	// 用于实现搜索建议功能
	FieldTypeCompletion = "completion"
)

// ===== Token 计数类型 (Token Count Type) =====

const (
	// FieldTypeTokenCount Token 计数字段
	// 存储分析后的 token 数量
	FieldTypeTokenCount = "token_count"
)

// ===== 其他特殊类型 (Other Types) =====

const (
	// FieldTypeAlias 字段别名
	FieldTypeAlias = "alias"

	// FieldTypeJoin 父子关系字段
	FieldTypeJoin = "join"

	// FieldTypePercolator 反向查询字段
	FieldTypePercolator = "percolator"

	// FieldTypeRankFeature 排名特征字段（用于排序）
	FieldTypeRankFeature = "rank_feature"

	// FieldTypeRankFeatures 排名特征集合
	FieldTypeRankFeatures = "rank_features"

	// FieldTypeDenseVector 密集向量（用于向量搜索）
	FieldTypeDenseVector = "dense_vector"

	// FieldTypeSparseVector 稀疏向量
	FieldTypeSparseVector = "sparse_vector"

	// FieldTypeSearchAsYouType 搜索即输入字段
	FieldTypeSearchAsYouType = "search_as_you_type"

	// FieldTypeHistogram 直方图聚合预聚合字段
	FieldTypeHistogram = "histogram"
)

// ===== 使用示例 =====

// 使用示例：
//
// 1. 创建带各种字段类型的索引：
//   builder.NewIndexBuilder(client, "products").
//       AddProperty("name", builder.FieldTypeText).               // 全文搜索字段
//       AddProperty("sku", builder.FieldTypeKeyword).             // 精确匹配字段
//       AddProperty("price", builder.FieldTypeFloat).             // 浮点数
//       AddProperty("quantity", builder.FieldTypeInt).            // 整数
//       AddProperty("available", builder.FieldTypeBoolean).       // 布尔值
//       AddProperty("created_at", builder.FieldTypeDate).         // 日期
//       AddProperty("location", builder.FieldTypeGeoPoint).       // 地理位置
//       AddProperty("tags", builder.FieldTypeKeyword).            // 标签数组
//       Create(ctx)
//
// 2. 嵌套对象：
//   builder.NewIndexBuilder(client, "articles").
//       AddProperty("author", builder.FieldTypeObject,
//           builder.WithSubProperties("name", builder.FieldTypeText),
//           builder.WithSubProperties("email", builder.FieldTypeKeyword),
//       ).
//       Create(ctx)
//
// 3. 数组类型字段：
//   builder.NewIndexBuilder(client, "posts").
//       AddProperty("comments", builder.FieldTypeNested,
//           builder.WithSubProperties("user", builder.FieldTypeKeyword),
//           builder.WithSubProperties("message", builder.FieldTypeText),
//           builder.WithSubProperties("date", builder.FieldTypeDate),
//       ).
//       Create(ctx)
