package builder

import "net/http"

// baseBuilder 基础构建器，提供通用的链式调用方法
type baseBuilder struct {
	headers http.Header
}

func (b *baseBuilder) initBaseBuilder() {
	b.headers = make(http.Header)
}

// Header 为单个请求设置自定义 HTTP Header
func (b *baseBuilder) Header(key, value string) {
	if b.headers == nil {
		b.headers = make(http.Header)
	}
	b.headers.Set(key, value)
}

func (b *baseBuilder) getHeaders() http.Header {
	return b.headers
}
