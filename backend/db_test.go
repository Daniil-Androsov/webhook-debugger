package backend

import (
	"testing"
	"time"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := NewDBFromPath(":memory:")
	if err != nil {
		t.Fatalf("NewDBFromPath: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSaveAndGetRequest(t *testing.T) {
	db := newTestDB(t)

	req := &Request{
		Method:    "POST",
		Path:      "/webhook",
		Headers:   map[string]string{"Content-Type": "application/json"},
		Body:      `{"key":"value"}`,
		Status:    200,
		Source:    "127.0.0.1:12345",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	id, err := db.SaveRequest(req)
	if err != nil {
		t.Fatalf("SaveRequest: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	got, err := db.GetRequest(id)
	if err != nil {
		t.Fatalf("GetRequest: %v", err)
	}

	if got.Method != req.Method {
		t.Errorf("method: got %q want %q", got.Method, req.Method)
	}
	if got.Path != req.Path {
		t.Errorf("path: got %q want %q", got.Path, req.Path)
	}
	if got.Body != req.Body {
		t.Errorf("body: got %q want %q", got.Body, req.Body)
	}
	if got.Status != req.Status {
		t.Errorf("status: got %d want %d", got.Status, req.Status)
	}
	if got.Source != req.Source {
		t.Errorf("source: got %q want %q", got.Source, req.Source)
	}
	if got.Headers["Content-Type"] != req.Headers["Content-Type"] {
		t.Errorf("header Content-Type: got %q want %q", got.Headers["Content-Type"], req.Headers["Content-Type"])
	}
}

func TestGetRequests(t *testing.T) {
	db := newTestDB(t)

	for i := 0; i < 3; i++ {
		_, err := db.SaveRequest(&Request{
			Method:    "GET",
			Path:      "/test",
			Headers:   map[string]string{},
			CreatedAt: time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("SaveRequest %d: %v", i, err)
		}
	}

	requests, err := db.GetRequests(2)
	if err != nil {
		t.Fatalf("GetRequests: %v", err)
	}
	if len(requests) != 2 {
		t.Errorf("got %d requests, want 2", len(requests))
	}
	// GetRequests orders by id DESC — first result is the newest
	if requests[0].ID <= requests[1].ID {
		t.Errorf("expected descending order: %d > %d", requests[0].ID, requests[1].ID)
	}
}

func TestUpdateStatus(t *testing.T) {
	db := newTestDB(t)

	id, err := db.SaveRequest(&Request{
		Method:    "POST",
		Path:      "/hook",
		Headers:   map[string]string{},
		Status:    0,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("SaveRequest: %v", err)
	}

	if err := db.UpdateStatus(id, 200); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, err := db.GetRequest(id)
	if err != nil {
		t.Fatalf("GetRequest: %v", err)
	}
	if got.Status != 200 {
		t.Errorf("status: got %d want 200", got.Status)
	}
}
