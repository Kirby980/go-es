package builder_test

import (
	"encoding/json"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

// --- Test types ---

type NestedItem struct {
	Name  string `es:"type:keyword"`
	Value int    `es:"type:integer"`
}

type DeeplyNestedChild struct {
	Label string `es:"type:keyword"`
}

type DeeplyNestedParent struct {
	Child DeeplyNestedChild `es:"type:object"`
}

// --- AutoMigrate tests ---

func TestAutoMigrate_SliceOfStruct(t *testing.T) {
	type Order struct {
		Title string       `es:"type:text"`
		Items []NestedItem `es:"type:nested"`
	}

	mock := &mockESClient{}
	b := builder.NewIndexBuilder(mock, "orders").AutoMigrate(Order{})
	body := b.Build()

	mappings, ok := body["mappings"].(map[string]any)
	if !ok {
		t.Fatal("expected mappings in build output")
	}
	properties, ok := mappings["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties in mappings")
	}

	items, ok := properties["items"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'items' in properties, got %v", properties)
	}

	if items["type"] != "nested" {
		t.Errorf("expected items type=nested, got %v", items["type"])
	}

	nestedProps, ok := items["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested properties for items, got %v", items)
	}

	nameField, ok := nestedProps["name"].(map[string]any)
	if !ok {
		t.Fatal("expected 'name' in nested properties")
	}
	if nameField["type"] != "keyword" {
		t.Errorf("expected name type=keyword, got %v", nameField["type"])
	}

	valueField, ok := nestedProps["value"].(map[string]any)
	if !ok {
		t.Fatal("expected 'value' in nested properties")
	}
	if valueField["type"] != "integer" {
		t.Errorf("expected value type=integer, got %v", valueField["type"])
	}
}

func TestAutoMigrate_SliceOfPointerStruct(t *testing.T) {
	type Order struct {
		Title string        `es:"type:text"`
		Items []*NestedItem `es:"type:nested"`
	}

	mock := &mockESClient{}
	b := builder.NewIndexBuilder(mock, "orders").AutoMigrate(Order{})
	body := b.Build()

	mappings := body["mappings"].(map[string]any)
	properties := mappings["properties"].(map[string]any)
	items := properties["items"].(map[string]any)

	if items["type"] != "nested" {
		t.Errorf("expected items type=nested, got %v", items["type"])
	}

	nestedProps, ok := items["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested properties for []*NestedItem, got %v", items)
	}

	if _, ok := nestedProps["name"]; !ok {
		t.Error("expected 'name' field in nested properties for []*NestedItem")
	}
	if _, ok := nestedProps["value"]; !ok {
		t.Error("expected 'value' field in nested properties for []*NestedItem")
	}
}

func TestAutoMigrate_PointerToSlice(t *testing.T) {
	type Order struct {
		Title string        `es:"type:text"`
		Items *[]NestedItem `es:"type:nested"`
	}

	mock := &mockESClient{}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AutoMigrate panicked on *[]NestedItem: %v", r)
		}
	}()

	b := builder.NewIndexBuilder(mock, "orders").AutoMigrate(Order{})
	body := b.Build()

	mappings := body["mappings"].(map[string]any)
	properties := mappings["properties"].(map[string]any)
	items := properties["items"].(map[string]any)

	if items["type"] != "nested" {
		t.Errorf("expected items type=nested, got %v", items["type"])
	}

	// Verify properties are present (Ptr -> Slice -> Struct path)
	nestedProps, ok := items["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested properties for *[]NestedItem, got %v", items)
	}
	if _, ok := nestedProps["name"]; !ok {
		t.Error("expected 'name' field in nested properties for *[]NestedItem")
	}
}

func TestAutoMigrate_DeeplyNested(t *testing.T) {
	type Root struct {
		Parent DeeplyNestedParent `es:"type:object"`
	}

	mock := &mockESClient{}
	b := builder.NewIndexBuilder(mock, "deep").AutoMigrate(Root{})
	body := b.Build()

	mappings := body["mappings"].(map[string]any)
	properties := mappings["properties"].(map[string]any)

	parent := properties["parent"].(map[string]any)
	if parent["type"] != "object" {
		t.Errorf("expected parent type=object, got %v", parent["type"])
	}

	parentProps := parent["properties"].(map[string]any)
	child := parentProps["child"].(map[string]any)
	if child["type"] != "object" {
		t.Errorf("expected child type=object, got %v", child["type"])
	}

	childProps := child["properties"].(map[string]any)
	label := childProps["label"].(map[string]any)
	if label["type"] != "keyword" {
		t.Errorf("expected label type=keyword, got %v", label["type"])
	}
}

func TestAutoMigrate_BasicTypes(t *testing.T) {
	type Product struct {
		Name        string  `es:"type:text"`
		Price       float64 `es:"type:float"`
		Quantity    int     `es:"type:integer"`
		Category    string  `es:"type:keyword"`
		Description string  `es:"type:text;analyzer:ik_smart"`
	}

	mock := &mockESClient{}
	b := builder.NewIndexBuilder(mock, "products").AutoMigrate(Product{})
	body := b.Build()

	mappings := body["mappings"].(map[string]any)
	properties := mappings["properties"].(map[string]any)

	expected := map[string]string{
		"name":        "text",
		"price":       "float",
		"quantity":    "integer",
		"category":    "keyword",
		"description": "text",
	}

	for field, expectedType := range expected {
		fieldMap, ok := properties[field].(map[string]any)
		if !ok {
			t.Errorf("expected field %q in properties", field)
			continue
		}
		if fieldMap["type"] != expectedType {
			t.Errorf("field %q: expected type=%q, got %v", field, expectedType, fieldMap["type"])
		}
	}

	// Verify analyzer on description
	desc := properties["description"].(map[string]any)
	if desc["analyzer"] != "ik_smart" {
		t.Errorf("expected description analyzer=ik_smart, got %v", desc["analyzer"])
	}
}

func TestAutoMigrate_FieldNameSnakeCase(t *testing.T) {
	// Verify that field names pass through toSnakeCase via AutoMigrate
	type LogEntry struct {
		HTTPMethod string `es:"type:keyword"`
		UserAgent  string `es:"type:text"`
	}

	mock := &mockESClient{}
	b := builder.NewIndexBuilder(mock, "logs").AutoMigrate(LogEntry{})
	body := b.Build()

	bodyJSON, _ := json.Marshal(body)
	bodyStr := string(bodyJSON)

	// HTTPMethod -> http_method (not h_t_t_p_method)
	if _, ok := body["mappings"].(map[string]any)["properties"].(map[string]any)["http_method"]; !ok {
		t.Errorf("expected field 'http_method' from HTTPMethod, got body: %s", bodyStr)
	}

	// UserAgent -> user_agent
	if _, ok := body["mappings"].(map[string]any)["properties"].(map[string]any)["user_agent"]; !ok {
		t.Errorf("expected field 'user_agent' from UserAgent, got body: %s", bodyStr)
	}
}
