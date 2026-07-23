package backend

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	db         *DB
	forwardURL string
	httpServer *http.Server
	onRequest  func(Request)
}

func NewServer(db *DB, onRequest func(Request)) *Server {
	return &Server{db: db, onRequest: onRequest}
}

func (s *Server) SetForwardURL(url string) {
	s.forwardURL = url
}

func (s *Server) Start() error {
	return s.StartOn(":9000")
}

func (s *Server) StartOn(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go s.httpServer.ListenAndServe()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)

	headers := make(map[string]string, len(r.Header))
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}

	req := &Request{
		Method:    r.Method,
		Path:      r.RequestURI,
		Headers:   headers,
		Body:      string(bodyBytes),
		Source:    r.RemoteAddr,
		CreatedAt: time.Now().Local(),
	}

	status := http.StatusOK

	if s.forwardURL != "" {
		status = s.forward(req, bodyBytes)
	}

	req.Status = status
	id, _ := s.db.SaveRequest(req)
	req.ID = id

	if s.onRequest != nil {
		s.onRequest(*req)
	}

	w.WriteHeader(status)
}

func (s *Server) forward(req *Request, body []byte) int {
	targetURL := s.forwardURL + req.Path

	fwdReq, err := http.NewRequest(req.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return http.StatusBadGateway
	}

	for k, v := range req.Headers {
		fwdReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(fwdReq)
	if err != nil {
		return http.StatusBadGateway
	}
	defer resp.Body.Close()

	return resp.StatusCode
}
