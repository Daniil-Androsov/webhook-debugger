package backend

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return fmt.Sprintf(":%d", port)
}

func startTestServer(t *testing.T, db *DB, onReq func(Request)) (srv *Server, baseURL string) {
	t.Helper()
	addr := freePort(t)
	srv = NewServer(db, onReq)
	if err := srv.StartOn(addr); err != nil {
		t.Fatalf("StartOn: %v", err)
	}
	// give the server a moment to bind
	time.Sleep(20 * time.Millisecond)
	t.Cleanup(func() {
		srv.Stop(nil)
	})
	return srv, "http://127.0.0.1" + addr
}

func TestServerCaptures(t *testing.T) {
	db := newTestDB(t)
	srv, base := startTestServer(t, db, nil)
	_ = srv

	body := `{"event":"test"}`
	resp, err := http.Post(base+"/webhook/path", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d want 200", resp.StatusCode)
	}

	requests, err := db.GetRequests(10)
	if err != nil {
		t.Fatalf("GetRequests: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("got %d requests, want 1", len(requests))
	}

	r := requests[0]
	if r.Method != "POST" {
		t.Errorf("method: got %q want POST", r.Method)
	}
	if r.Path != "/webhook/path" {
		t.Errorf("path: got %q want /webhook/path", r.Path)
	}
	if r.Body != body {
		t.Errorf("body: got %q want %q", r.Body, body)
	}
	if r.Headers["Content-Type"] == "" {
		t.Error("Content-Type header not captured")
	}
}

func TestServerForwards(t *testing.T) {
	received := make(chan string, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received <- string(b)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer target.Close()

	db := newTestDB(t)
	srv, base := startTestServer(t, db, nil)
	srv.SetForwardURL(target.URL)

	body := `{"forwarded":true}`
	resp, err := http.Post(base+"/fwd", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status: got %d want 202", resp.StatusCode)
	}

	select {
	case got := <-received:
		if got != body {
			t.Errorf("target body: got %q want %q", got, body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("target did not receive request")
	}
}

func TestServerNoForward(t *testing.T) {
	db := newTestDB(t)
	_, base := startTestServer(t, db, nil)

	resp, err := http.Get(base + "/ping")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d want 200", resp.StatusCode)
	}
}
