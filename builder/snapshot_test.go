package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestSnapshotBuilder_CreateRepository(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPut {
				t.Errorf("Expected PUT, got %s", method)
			}
			if path != "/_snapshot/my_repo" {
				t.Errorf("Expected path /_snapshot/my_repo, got %s", path)
			}

			resp := `{"acknowledged": true}`
			return []byte(resp), nil
		},
	}

	b := builder.NewSnapshotBuilder(mock).
		Repository("my_repo").
		Settings(map[string]any{"location": "/mount/backups"}).
		Debug(true)

	resp, err := b.CreateRepository(context.Background(), "fs")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !resp.Acknowledged {
		t.Errorf("Expected acknowledged true")
	}
}

func TestSnapshotBuilder_Create(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPut {
				t.Errorf("Expected PUT, got %s", method)
			}
			if path != "/_snapshot/my_repo/my_snapshot" {
				t.Errorf("Expected path /_snapshot/my_repo/my_snapshot, got %s", path)
			}

			resp := `{"acknowledged": true}`
			return []byte(resp), nil
		},
	}

	b := builder.NewSnapshotBuilder(mock).
		Repository("my_repo").
		Snapshot("my_snapshot").
		Indices("index1", "index2").
		Debug(true)

	resp, err := b.Create(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !resp.Acknowledged {
		t.Errorf("Expected acknowledged true")
	}
}

func TestSnapshotBuilder_Restore(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPost {
				t.Errorf("Expected POST, got %s", method)
			}
			if path != "/_snapshot/my_repo/my_snapshot/_restore" {
				t.Errorf("Expected path /_snapshot/my_repo/my_snapshot/_restore, got %s", path)
			}

			resp := `{"acknowledged": true}`
			return []byte(resp), nil
		},
	}

	b := builder.NewSnapshotBuilder(mock).
		Repository("my_repo").
		Snapshot("my_snapshot").
		Debug(true)

	resp, err := b.Restore(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !resp.Acknowledged {
		t.Errorf("Expected acknowledged true")
	}
}
