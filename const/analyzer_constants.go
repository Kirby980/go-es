package constant

// ===== 分析器类型 (Analyzer Types) =====

const (
	AnalyzerTypeCustom     = "custom"     // 自定义分析器
	AnalyzerTypeStandard   = "standard"   // 标准分析器（默认）
	AnalyzerTypeSimple     = "simple"     // 简单分析器（按非字母字符分词）
	AnalyzerTypeWhitespace = "whitespace" // 空格分析器（按空格分词）
	AnalyzerTypeKeyword    = "keyword"    // 关键词分析器（不分词）
	AnalyzerTypePattern    = "pattern"    // 正则分析器
	AnalyzerTypeLanguage   = "language"   // 语言分析器
)

// IK 分词器内置分析器
const (
	AnalyzerIKSmart   = "ik_smart"    // IK智能分词（粗粒度，推荐用于搜索）
	AnalyzerIKMaxWord = "ik_max_word" // IK最大词数分词（细粒度，推荐用于索引）
)

// ===== 分词器 (Tokenizers) =====

const (
	// 标准分词器
	TokenizerStandard   = "standard"   // 标准分词器（基于Unicode文本分段）
	TokenizerLetter     = "letter"     // 字母分词器（按非字母字符分词）
	TokenizerLowercase  = "lowercase"  // 小写分词器（按非字母分词并转小写）
	TokenizerWhitespace = "whitespace" // 空格分词器（按空格分词）
	TokenizerKeyword    = "keyword"    // 关键词分词器（不分词，整个文本作为一个token）
	TokenizerPattern    = "pattern"    // 正则分词器

	// IK 中文分词器
	TokenizerIKSmart   = "ik_smart"    // IK智能分词（粗粒度）："中华人民共和国" → ["中华人民共和国"]
	TokenizerIKMaxWord = "ik_max_word" // IK最大词数分词（细粒度）："中华人民共和国" → ["中华人民共和国","中华人民","中华","华人","人民共和国","人民","共和国","共和","国"]

	// N-Gram 分词器
	TokenizerNGram     = "ngram"      // N-Gram分词器（滑动窗口分词）
	TokenizerEdgeNGram = "edge_ngram" // Edge N-Gram分词器（从开头开始的滑动窗口）

	// 路径分词器
	TokenizerPathHierarchy = "path_hierarchy" // 路径层次分词器："/foo/bar/baz" → ["/foo", "/foo/bar", "/foo/bar/baz"]

	// UAX URL Email 分词器
	TokenizerUAXUrlEmail = "uax_url_email" // 将URL和Email作为单个token
)

// ===== 字符过滤器 (Char Filters) - 在分词前对原始文本进行处理 =====

const (
	CharFilterHTMLStrip = "html_strip"      // 去除HTML标签：<p>Hello</p> → Hello
	CharFilterMapping   = "mapping"         // 字符映射替换：& → and
	CharFilterPattern   = "pattern_replace" // 正则替换
)

// ===== Token 过滤器 (Token Filters) - 在分词后对tokens进行处理 =====

const (
	// 大小写转换
	TokenFilterLowercase = "lowercase" // 转小写：Hello → hello
	TokenFilterUppercase = "uppercase" // 转大写：hello → HELLO

	// 停用词过滤
	TokenFilterStop = "stop" // 去除停用词（the, is, at等）

	// 词干提取
	TokenFilterStemmer    = "stemmer"     // 词干提取：running → run
	TokenFilterPorterStem = "porter_stem" // Porter词干算法
	TokenFilterSnowball   = "snowball"    // Snowball词干算法
	TokenFilterKStem      = "kstem"       // KStem词干算法

	// 同义词
	TokenFilterSynonym      = "synonym"       // 同义词替换：quick → [quick, fast]
	TokenFilterSynonymGraph = "synonym_graph" // 同义词图（支持多词同义词）

	// 去重
	TokenFilterUnique = "unique" // 去除重复的token

	// 截断
	TokenFilterTruncate = "truncate" // 截断token到指定长度
	TokenFilterLimit    = "limit"    // 限制token数量

	// 反转
	TokenFilterReverse = "reverse" // 反转token：hello → olleh

	// 长度过滤
	TokenFilterLength = "length" // 过滤指定长度范围的token

	// N-Gram
	TokenFilterNGram     = "ngram"      // N-Gram过滤
	TokenFilterEdgeNGram = "edge_ngram" // Edge N-Gram过滤

	// ASCIIFolding（将Unicode字符转为ASCII）
	TokenFilterASCIIFolding = "asciifolding" // café → cafe

	// 词形还原
	TokenFilterLemmatizer = "lemmatizer" // 词形还原：better → good

	// 拼音（需要插件）
	TokenFilterPinyin = "pinyin" // 中文转拼音（需要 elasticsearch-analysis-pinyin 插件）

	// 去除标点
	TokenFilterTrim = "trim" // 去除token首尾空格

	// 单词分割
	TokenFilterWordDelimiter      = "word_delimiter"       // 单词分隔符：WiFi → [Wi, Fi]
	TokenFilterWordDelimiterGraph = "word_delimiter_graph" // 单词分隔符图
)

// ===== 常用组合示例 =====

// 使用示例：
//
// 1. 中文分词（保持大小写）：
//   AddCustomAnalyzer("ik_case_sensitive", TokenizerIKSmart)
//
// 2. 中文分词（转小写）：
//   AddCustomAnalyzer("ik_lowercase", TokenizerIKSmart, TokenFilterLowercase)
//
// 3. 去除HTML后进行IK分词：
//   AddAnalyzer("html_ik",
//       WithAnalyzerType(AnalyzerTypeCustom),
//       WithCharFilters(CharFilterHTMLStrip),
//       WithTokenizer(TokenizerIKSmart),
//   )
//
// 4. 英文分词（词干提取+小写+去停用词）：
//   AddAnalyzer("english_stemmed",
//       WithAnalyzerType(AnalyzerTypeCustom),
//       WithTokenizer(TokenizerStandard),
//       WithTokenFilters(TokenFilterLowercase, TokenFilterStop, TokenFilterPorterStem),
//   )
