package builder_test

import (
	"context"
	"net/http"

	"github.com/Kirby980/go-es/logger"
)

// mockESClient is a reusable mock implementing builder.ESClient for unit tests.
// It captures the last method, path, and body from Do() calls and delegates to
// a configurable doFunc for response/error control.
type mockESClient struct {
	doFunc     func(ctx context.Context, method, path string, body any) ([]byte, error)
	doReqFunc  func(ctx context.Context, req *http.Request) ([]byte, error)
	lastMethod string
	lastPath   string
	lastBody   any
	address    string
}

func (m *mockESClient) Do(ctx context.Context, method, path string, body any) ([]byte, error) {
	m.lastMethod = method
	m.lastPath = path
	m.lastBody = body
	if m.doFunc != nil {
		return m.doFunc(ctx, method, path, body)
	}
	return []byte(`{}`), nil
}

func (m *mockESClient) DoWithHeader(ctx context.Context, method, path string, body any, header http.Header) ([]byte, error) {
	m.lastMethod = method
	m.lastPath = path
	m.lastBody = body
	if m.doFunc != nil {
		return m.doFunc(ctx, method, path, body)
	}
	return []byte(`{}`), nil
}

func (m *mockESClient) GetAddress() string {
	if m.address != "" {
		return m.address
	}
	return "http://localhost:9200"
}

func (m *mockESClient) DoRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	if m.doReqFunc != nil {
		return m.doReqFunc(ctx, req)
	}
	return []byte(`{}`), nil
}

func (m *mockESClient) GetLogger() logger.Logger {
	return logger.NopLogger{}
}
