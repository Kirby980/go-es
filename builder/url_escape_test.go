package builder_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestDocumentBuilder_URLEscaping(t *testing.T) {
	tests := []struct {
		name           string
		index          string
		id             string
		mustNotContain []string // raw chars that must NOT appear unescaped in path
		mustContain    []string // escaped forms that MUST appear in path
	}{
		{
			name:           "slash in index and id",
			index:          "my/index",
			id:             "doc/1",
			mustNotContain: nil, // slash is in the path separator too, check via mustContain
			mustContain:    []string{"my%2Findex", "doc%2F1"},
		},
		{
			name:           "hash in id",
			index:          "test-index",
			id:             "doc#1",
			mustNotContain: []string{"#"},
			mustContain:    []string{"doc%231"},
		},
		{
			name:           "space in index",
			index:          "test index",
			id:             "doc-1",
			mustNotContain: []string{" "},
			mustContain:    []string{"test%20index"},
		},
		{
			name:           "question mark in id",
			index:          "test-index",
			id:             "id?query",
			mustNotContain: []string{"?"},
			mustContain:    []string{"id%3Fquery"},
		},
		{
			name:        "non-ASCII in index",
			index:       "\u4e2d\u6587\u7d22\u5f15", // Chinese characters
			id:          "doc-1",
			mustContain: []string{"%"}, // non-ASCII must be percent-encoded
		},
		{
			name:        "slashes in id",
			index:       "test-index",
			id:          "id/with/slashes",
			mustContain: []string{"id%2Fwith%2Fslashes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockESClient{
				doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
					return []byte(`{"_index":"x","_id":"1","_version":1,"result":"created","_shards":{"total":1,"successful":1,"failed":0}}`), nil
				},
			}

			// Use Do() which calls url.PathEscape on index and id
			_, _ = builder.NewDocumentBuilder(mock, tt.index).ID(tt.id).Set("f", "v").Do(context.Background())

			path := mock.lastPath
			if path == "" {
				t.Fatal("expected a path to be captured, got empty string")
			}

			for _, banned := range tt.mustNotContain {
				if strings.Contains(path, banned) {
					t.Errorf("path %q contains raw %q (should be escaped)", path, banned)
				}
			}

			for _, expected := range tt.mustContain {
				if !strings.Contains(path, expected) {
					t.Errorf("path %q does not contain expected escaped form %q", path, expected)
				}
			}
		})
	}
}

func TestSearchBuilder_URLEscaping(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return []byte(`{"took":0,"timed_out":false,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":0,"relation":"eq"},"max_score":0,"hits":[]}}`), nil
		},
	}

	_, _ = builder.NewSearchBuilder(mock, "my/special index").
		MatchAll().
		Do(context.Background())

	path := mock.lastPath
	if path == "" {
		t.Fatal("expected a path, got empty")
	}

	if strings.Contains(path, " ") {
		t.Errorf("path %q contains unescaped space", path)
	}

	if !strings.Contains(path, "my%2Fspecial%20index") {
		t.Errorf("path %q does not contain properly escaped index name", path)
	}
}

func TestIndexBuilder_URLEscaping(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return []byte(`{}`), nil
		},
	}

	_, _ = builder.NewIndexBuilder(mock, "my/index#name").Exists(context.Background())

	path := mock.lastPath
	if path == "" {
		t.Fatal("expected a path, got empty")
	}

	if strings.Contains(path, "#") {
		t.Errorf("path %q contains unescaped #", path)
	}

	if !strings.Contains(path, "my%2Findex%23name") {
		t.Errorf("path %q does not contain properly escaped index name", path)
	}
}
