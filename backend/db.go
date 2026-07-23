package backend

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Request struct {
	ID        int               `json:"id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	Status    int               `json:"status"`
	Source    string            `json:"source"`
	CreatedAt time.Time         `json:"created_at"`
}

type DB struct {
	conn *sql.DB
}

func NewDB() (*DB, error) {
	dir := filepath.Join(os.Getenv("HOME"), ".webhook-debugger")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return NewDBFromPath(filepath.Join(dir, "requests.db"))
}

func NewDBFromPath(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := migrate(conn); err != nil {
		return nil, err
	}
	return &DB{conn: conn}, nil
}

func migrate(conn *sql.DB) error {
	_, err := conn.Exec(`CREATE TABLE IF NOT EXISTS requests (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		method     TEXT NOT NULL,
		path       TEXT NOT NULL,
		headers    TEXT NOT NULL DEFAULT '{}',
		body       TEXT NOT NULL DEFAULT '',
		status     INTEGER NOT NULL DEFAULT 0,
		source     TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

func (d *DB) SaveRequest(r *Request) (int, error) {
	headers, err := json.Marshal(r.Headers)
	if err != nil {
		return 0, err
	}

	res, err := d.conn.Exec(
		`INSERT INTO requests (method, path, headers, body, status, source, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.Method, r.Path, string(headers), r.Body, r.Status, r.Source, r.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return int(id), err
}

func (d *DB) UpdateStatus(id, status int) error {
	_, err := d.conn.Exec(`UPDATE requests SET status = ? WHERE id = ?`, status, id)
	return err
}

func (d *DB) GetRequests(limit int) ([]Request, error) {
	rows, err := d.conn.Query(
		`SELECT id, method, path, headers, body, status, source, created_at FROM requests ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		r, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

func (d *DB) GetRequest(id int) (Request, error) {
	row := d.conn.QueryRow(
		`SELECT id, method, path, headers, body, status, source, created_at FROM requests WHERE id = ?`,
		id,
	)
	return scanRequest(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRequest(s scanner) (Request, error) {
	var r Request
	var headersJSON string
	var createdAt string

	err := s.Scan(&r.ID, &r.Method, &r.Path, &headersJSON, &r.Body, &r.Status, &r.Source, &createdAt)
	if err != nil {
		return r, err
	}

	if err := json.Unmarshal([]byte(headersJSON), &r.Headers); err != nil {
		r.Headers = map[string]string{}
	}

	r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return r, nil
}

func (d *DB) DeleteRequest(id int) error {
	_, err := d.conn.Exec(`DELETE FROM requests WHERE id = ?`, id)
	return err
}

func (d *DB) ClearRequests() error {
	_, err := d.conn.Exec(`DELETE FROM requests`)
	return err
}

func (d *DB) Close() error {
	return d.conn.Close()
}
