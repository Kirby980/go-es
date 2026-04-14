package builder_test

import (
	"encoding/json"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

// --- MatchAll Bug fix test ---

func TestSearchBuilder_MatchAll_Build(t *testing.T) {
	mock := &mockESClient{}
	b := builder.NewSearchBuilder(mock, "test-index")
	body := b.MatchAll().Build()

	query, ok := body["query"]
	if !ok {
		t.Fatal("MatchAll().Build() should include a 'query' key")
	}

	queryMap, ok := query.(map[string]any)
	if !ok {
		t.Fatalf("query should be map[string]any, got %T", query)
	}

	if _, ok := queryMap["match_all"]; !ok {
		t.Errorf("query should contain 'match_all', got %v", queryMap)
	}
}

func TestSearchBuilder_MatchAll_NotOverriddenByEmptyBool(t *testing.T) {
	mock := &mockESClient{}
	// MatchAll without any bool conditions should use match_all
	body := builder.NewSearchBuilder(mock, "test-index").MatchAll().Build()

	data, _ := json.Marshal(body["query"])
	if string(data) == "" || string(data) == "null" {
		t.Fatal("query should not be empty when MatchAll is set")
	}

	var q map[string]any
	json.Unmarshal(data, &q)
	if _, ok := q["match_all"]; !ok {
		t.Errorf("expected match_all query, got: %s", string(data))
	}
}

func TestSearchBuilder_BoolOverridesMatchAll(t *testing.T) {
	mock := &mockESClient{}
	// When bool conditions exist, they should take priority over match_all
	body := builder.NewSearchBuilder(mock, "test-index").
		MatchAll().
		Term("status", "active").
		Build()

	query, ok := body["query"].(map[string]any)
	if !ok {
		t.Fatal("query should exist")
	}

	// Bool query should take priority
	if _, ok := query["bool"]; !ok {
		t.Error("bool query should override match_all when bool conditions exist")
	}
}

// --- Scan method tests for ScrollResponse and SearchAfterResponse ---

func TestScrollResponse_Scan(t *testing.T) {
	resp := &builder.ScrollResponse{}
	resp.Hits.Hits = []builder.HitItem{
		{Source: map[string]any{"name": "alice", "age": float64(30)}},
		{Source: map[string]any{"name": "bob", "age": float64(25)}},
	}
	resp.Hits.Total = builder.HitsTotal{Value: 2, Relation: "eq"}

	type Person struct {
		Name string  `json:"name"`
		Age  float64 `json:"age"`
	}
	var people []Person
	if err := resp.Scan(&people); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(people) != 2 {
		t.Fatalf("expected 2 results, got %d", len(people))
	}
	if people[0].Name != "alice" {
		t.Errorf("expected 'alice', got %q", people[0].Name)
	}
}

func TestScrollResponse_Total(t *testing.T) {
	resp := &builder.ScrollResponse{}
	resp.Hits.Total = builder.HitsTotal{Value: 42, Relation: "eq"}
	if resp.Total() != 42 {
		t.Errorf("expected 42, got %d", resp.Total())
	}
}

func TestSearchAfterResponse_Scan(t *testing.T) {
	resp := &builder.SearchAfterResponse{}
	resp.Hits.Hits = []builder.HitItem{
		{Source: map[string]any{"title": "doc1"}, Sort: []any{"a"}},
		{Source: map[string]any{"title": "doc2"}, Sort: []any{"b"}},
	}
	resp.Hits.Total = builder.HitsTotal{Value: 2, Relation: "eq"}

	type Doc struct {
		Title string `json:"title"`
	}
	var docs []Doc
	if err := resp.Scan(&docs); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected 2 results, got %d", len(docs))
	}
	if docs[0].Title != "doc1" {
		t.Errorf("expected 'doc1', got %q", docs[0].Title)
	}
}

func TestSearchAfterResponse_Total(t *testing.T) {
	resp := &builder.SearchAfterResponse{}
	resp.Hits.Total = builder.HitsTotal{Value: 100, Relation: "gte"}
	if resp.Total() != 100 {
		t.Errorf("expected 100, got %d", resp.Total())
	}
}

func TestScan_InvalidDest(t *testing.T) {
	resp := &builder.ScrollResponse{}
	resp.Hits.Hits = []builder.HitItem{
		{Source: map[string]any{"name": "test"}},
	}

	// Non-pointer should fail
	var s []string
	if err := resp.Scan(s); err == nil {
		t.Error("expected error when dest is not a pointer")
	}

	// Pointer to non-slice should fail
	var n int
	if err := resp.Scan(&n); err == nil {
		t.Error("expected error when dest is not a pointer to slice")
	}
}
