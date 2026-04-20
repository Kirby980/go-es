package builder

// BoolQuery 泛型布尔查询构建器，通过嵌入方式为各 Builder 提供公共的查询方法
type BoolQuery[T any] struct {
	self               *T
	filters            []map[string]any
	must               []map[string]any
	should             []map[string]any
	mustNot            []map[string]any
	minimumShouldMatch any
}

// initBoolQuery 初始化 BoolQuery，在外层 Builder 构造函数中调用
func (q *BoolQuery[T]) initBoolQuery(self *T) {
	q.self = self
	q.filters = make([]map[string]any, 0)
	q.must = make([]map[string]any, 0)
	q.should = make([]map[string]any, 0)
	q.mustNot = make([]map[string]any, 0)
}

// ========== Must 条件 ==========

// MustRaw 注入原生的 must 查询 DSL
func (q *BoolQuery[T]) MustRaw(query map[string]any) *T {
	q.must = append(q.must, query)
	return q.self
}

// Match 添加 match 查询
func (q *BoolQuery[T]) Match(field string, value any) *T {
	q.must = append(q.must, map[string]any{
		"match": map[string]any{
			field: value,
		},
	})
	return q.self
}

// MatchBoost 添加带 boost 的 match 查询
func (q *BoolQuery[T]) MatchBoost(field string, value any, boost float64) *T {
	q.must = append(q.must, map[string]any{
		"match": map[string]any{
			field: map[string]any{
				"query": value,
				"boost": boost,
			},
		},
	})
	return q.self
}

// MatchPhrase 添加 match_phrase 查询
func (q *BoolQuery[T]) MatchPhrase(field string, value any) *T {
	q.must = append(q.must, map[string]any{
		"match_phrase": map[string]any{
			field: value,
		},
	})
	return q.self
}

// ========== Filter 条件 ==========

// FilterRaw 注入原生的 filter 查询 DSL
func (q *BoolQuery[T]) FilterRaw(query map[string]any) *T {
	q.filters = append(q.filters, query)
	return q.self
}

// Term 添加 term 查询
func (q *BoolQuery[T]) Term(field string, value any) *T {
	q.filters = append(q.filters, map[string]any{
		"term": map[string]any{
			field: value,
		},
	})
	return q.self
}

// TermBoost 添加带 boost 的 term 查询
func (q *BoolQuery[T]) TermBoost(field string, value any, boost float64) *T {
	q.filters = append(q.filters, map[string]any{
		"term": map[string]any{
			field: map[string]any{
				"value": value,
				"boost": boost,
			},
		},
	})
	return q.self
}

// Terms 添加 terms 查询
func (q *BoolQuery[T]) Terms(field string, values ...any) *T {
	q.filters = append(q.filters, map[string]any{
		"terms": map[string]any{
			field: values,
		},
	})
	return q.self
}

// TermsBoost 添加带 boost 的 terms 查询
func (q *BoolQuery[T]) TermsBoost(field string, boost float64, values ...any) *T {
	q.filters = append(q.filters, map[string]any{
		"terms": map[string]any{
			field:   values,
			"boost": boost,
		},
	})
	return q.self
}

// Range 添加范围查询
func (q *BoolQuery[T]) Range(field string, gte, lte any) *T {
	rangeQuery := make(map[string]any)
	if gte != nil {
		rangeQuery["gte"] = gte
	}
	if lte != nil {
		rangeQuery["lte"] = lte
	}
	q.filters = append(q.filters, map[string]any{
		"range": map[string]any{
			field: rangeQuery,
		},
	})
	return q.self
}

// Exists 添加字段存在查询
func (q *BoolQuery[T]) Exists(field string) *T {
	q.filters = append(q.filters, map[string]any{
		"exists": map[string]any{
			"field": field,
		},
	})
	return q.self
}

// ========== Should 条件 ==========

// ShouldRaw 注入原生的 should 查询 DSL
func (q *BoolQuery[T]) ShouldRaw(query map[string]any) *T {
	q.should = append(q.should, query)
	return q.self
}

// MatchShould 添加 match 查询到 should 条件
func (q *BoolQuery[T]) MatchShould(field string, value any) *T {
	q.should = append(q.should, map[string]any{
		"match": map[string]any{
			field: value,
		},
	})
	return q.self
}

// MatchShouldBoost 添加带 boost 的 match 查询到 should 条件
func (q *BoolQuery[T]) MatchShouldBoost(field string, value any, boost float64) *T {
	q.should = append(q.should, map[string]any{
		"match": map[string]any{
			field: map[string]any{
				"query": value,
				"boost": boost,
			},
		},
	})
	return q.self
}

// TermShould 添加 term 查询到 should 条件
func (q *BoolQuery[T]) TermShould(field string, value any) *T {
	q.should = append(q.should, map[string]any{
		"term": map[string]any{
			field: value,
		},
	})
	return q.self
}

// TermShouldBoost 添加带 boost 的 term 查询到 should 条件
func (q *BoolQuery[T]) TermShouldBoost(field string, value any, boost float64) *T {
	q.should = append(q.should, map[string]any{
		"term": map[string]any{
			field: map[string]any{
				"value": value,
				"boost": boost,
			},
		},
	})
	return q.self
}

// TermsShould 添加 terms 查询到 should 条件
func (q *BoolQuery[T]) TermsShould(field string, values ...any) *T {
	q.should = append(q.should, map[string]any{
		"terms": map[string]any{
			field: values,
		},
	})
	return q.self
}

// TermsShouldBoost 添加带 boost 的 terms 查询到 should 条件
func (q *BoolQuery[T]) TermsShouldBoost(field string, boost float64, values ...any) *T {
	q.should = append(q.should, map[string]any{
		"terms": map[string]any{
			field:   values,
			"boost": boost,
		},
	})
	return q.self
}

// RangeShould 添加范围查询到 should 条件
func (q *BoolQuery[T]) RangeShould(field string, gte, lte any) *T {
	rangeQuery := make(map[string]any)
	if gte != nil {
		rangeQuery["gte"] = gte
	}
	if lte != nil {
		rangeQuery["lte"] = lte
	}
	q.should = append(q.should, map[string]any{
		"range": map[string]any{
			field: rangeQuery,
		},
	})
	return q.self
}

// ========== Must Not 条件 ==========

// MustNotRaw 注入原生的 must_not 查询 DSL
func (q *BoolQuery[T]) MustNotRaw(query map[string]any) *T {
	q.mustNot = append(q.mustNot, query)
	return q.self
}

// MustNot 添加 must_not 条件（term 查询）
func (q *BoolQuery[T]) MustNot(field string, value any) *T {
	q.mustNot = append(q.mustNot, map[string]any{
		"term": map[string]any{
			field: value,
		},
	})
	return q.self
}

// TermMustNot 添加 term 查询到 must_not 条件（等同于 MustNot，提供别名以保持一致性）
func (q *BoolQuery[T]) TermMustNot(field string, value any) *T {
	return q.MustNot(field, value)
}

// MatchMustNot 添加 match 查询到 must_not 条件
func (q *BoolQuery[T]) MatchMustNot(field string, value any) *T {
	q.mustNot = append(q.mustNot, map[string]any{
		"match": map[string]any{
			field: value,
		},
	})
	return q.self
}

// RangeMustNot 添加范围查询到 must_not 条件
func (q *BoolQuery[T]) RangeMustNot(field string, gte, lte any) *T {
	rangeQuery := make(map[string]any)
	if gte != nil {
		rangeQuery["gte"] = gte
	}
	if lte != nil {
		rangeQuery["lte"] = lte
	}
	q.mustNot = append(q.mustNot, map[string]any{
		"range": map[string]any{
			field: rangeQuery,
		},
	})
	return q.self
}

// ========== 配置 ==========

// MinimumShouldMatch 设置最少匹配 should 条件数量
// 参数可以是：
// - 整数：至少匹配的 should 条件数量，如 2 表示至少匹配2个条件
// - 字符串：百分比或表达式，如 "75%" 表示至少匹配75%的条件
func (q *BoolQuery[T]) MinimumShouldMatch(value any) *T {
	q.minimumShouldMatch = value
	return q.self
}

// ========== 构建 ==========

// buildBoolQuery 构建 bool 查询部分，返回可直接赋值给 body["query"] 的 map
// 无查询条件时返回 nil
func (q *BoolQuery[T]) buildBoolQuery() map[string]any {
	if len(q.must) == 0 && len(q.filters) == 0 && len(q.should) == 0 && len(q.mustNot) == 0 {
		return nil
	}

	boolQuery := make(map[string]any)
	if len(q.must) > 0 {
		boolQuery["must"] = q.must
	}
	if len(q.filters) > 0 {
		boolQuery["filter"] = q.filters
	}
	if len(q.should) > 0 {
		boolQuery["should"] = q.should
	}
	if len(q.mustNot) > 0 {
		boolQuery["must_not"] = q.mustNot
	}
	if q.minimumShouldMatch != nil {
		boolQuery["minimum_should_match"] = q.minimumShouldMatch
	}

	return map[string]any{
		"bool": boolQuery,
	}
}
