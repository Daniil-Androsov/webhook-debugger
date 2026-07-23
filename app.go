package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"webhook-debugger/backend"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	db        *backend.DB
	server    *backend.Server
	tunnel    *backend.Tunnel
	TunnelURL string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	db, err := backend.NewDB()
	if err != nil {
		runtime.LogErrorf(ctx, "DB init failed: %v", err)
		return
	}
	a.db = db

	a.server = backend.NewServer(db, func(r backend.Request) {
		runtime.EventsEmit(ctx, "new_request", r)
	})

	if err := a.server.Start(); err != nil {
		runtime.LogErrorf(ctx, "Server start failed: %v", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.tunnel != nil {
		a.tunnel.Stop()
	}
	if a.server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.server.Stop(shutdownCtx)
	}
	if a.db != nil {
		a.db.Close()
	}
}

func (a *App) GetRequests() []backend.Request {
	requests, err := a.db.GetRequests(500)
	if err != nil {
		runtime.LogErrorf(a.ctx, "GetRequests: %v", err)
		return nil
	}
	return requests
}

func (a *App) GetRequest(id int) backend.Request {
	r, err := a.db.GetRequest(id)
	if err != nil {
		runtime.LogErrorf(a.ctx, "GetRequest %d: %v", id, err)
		return backend.Request{}
	}
	return r
}

func (a *App) SetForwardURL(url string) {
	a.server.SetForwardURL(url)
}

func (a *App) DeleteRequest(id int) error {
	return a.db.DeleteRequest(id)
}

func (a *App) ClearRequests() error {
	if err := a.db.ClearRequests(); err != nil {
		return err
	}
	runtime.EventsEmit(a.ctx, "requests_cleared", nil)
	return nil
}

func (a *App) StartTunnel() (string, error) {
	if a.tunnel != nil {
		return a.TunnelURL, nil
	}

	t, err := backend.StartTunnel(9000)
	if err != nil {
		return "", err
	}

	a.tunnel = t
	a.TunnelURL = t.URL
	runtime.EventsEmit(a.ctx, "tunnel_started", t.URL)
	return t.URL, nil
}

func (a *App) StopTunnel() {
	if a.tunnel != nil {
		a.tunnel.Stop()
		a.tunnel = nil
		a.TunnelURL = ""
		runtime.EventsEmit(a.ctx, "tunnel_stopped", nil)
	}
}

func (a *App) ReplayRequest(id int, targetURL string) error {
	r, err := a.db.GetRequest(id)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(r.Method, targetURL+r.Path, bytes.NewBufferString(r.Body))
	if err != nil {
		return err
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (a *App) ExportAsCurl(id int) string {
	r, err := a.db.GetRequest(id)
	if err != nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "curl -X %s", r.Method)

	for k, v := range r.Headers {
		fmt.Fprintf(&b, " \\\n  -H '%s: %s'", k, v)
	}

	if r.Body != "" {
		fmt.Fprintf(&b, " \\\n  -d '%s'", strings.ReplaceAll(r.Body, "'", "'\\''"))
	}

	fmt.Fprintf(&b, " \\\n  'http://localhost:9000%s'", r.Path)

	return b.String()
}
