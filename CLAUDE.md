# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Webhook Debugger — a local-first desktop app for inspecting, replaying, and editing incoming webhook requests. Built with Wails v2 (Go backend + React frontend).

Core value: all data stays on the user's machine. No cloud logging, no third-party servers for request data.

## Stack

- **Wails v2** — desktop wrapper (Go + React in one binary)
- **Go 1.22** — HTTP server, tunnel client, SQLite access, replay logic
- **React + Tailwind** — UI (inside `frontend/`)
- **`modernc.org/sqlite`** — pure Go SQLite driver, no cgo required

## Commands

```bash
# Install Wails CLI (once)
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Run in dev mode (hot reload)
wails dev

# Build production binary
wails build

# Run Go tests
go test ./backend/...

# Run a single test
go test ./backend/... -run TestName
```

## Architecture

The app has three layers that work together:

**1. Receiver (`backend/server.go`)**
Local HTTP server on port 9000. Every incoming request is captured in full (method, path, headers, body) and written to SQLite before being forwarded to a user-configured target URL. This is the core of the app.

**2. Storage (`backend/db.go`)**
SQLite database at `~/.webhook-debugger/requests.db`. Schema: `requests` table with `id, method, path, headers (JSON text), body (text), status, source, created_at`. All reads/writes go through this layer.

**3. Wails Bindings (`backend/app.go`)**
Go structs exposed to the React frontend via Wails bindings. Methods like `GetRequests()`, `ReplayRequest()`, `ExportAsCurl()` are called directly from JS as `window.go.app.MethodName()`.

**Frontend (`frontend/src/`)**
React app. Communicates with Go exclusively through Wails bindings — no fetch/axios to localhost. Real-time updates via Wails events emitted from Go on each new request.

## Key design decisions

- **No cgo**: use `modernc.org/sqlite` so the binary compiles without a C toolchain.
- **Tunnel**: MVP uses cloudflared (launched as a subprocess) to create a public URL. The tunnel layer is isolated so it can be swapped later.
- **Replay**: sends an exact HTTP copy of the stored request to any target URL the user specifies. Headers and body are editable before sending.
- **curl export**: formats the stored request as a `curl` command string — high-value, low-effort feature for developers.
