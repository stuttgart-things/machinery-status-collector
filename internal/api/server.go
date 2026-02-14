package api

import (
	"net/http"

	"github.com/stuttgart-things/machinery-status-collector/internal/collector"
)

// Server holds the HTTP handler and its dependencies.
type Server struct {
	store   *collector.StatusStore
	version string
	commit  string
	Handler http.Handler
}

// NewServer creates a Server with all routes and middleware registered.
func NewServer(store *collector.StatusStore, version, commit string) *Server {
	s := &Server{
		store:   store,
		version: version,
		commit:  commit,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/status", s.handlePostStatus)
	mux.HandleFunc("GET /api/v1/status", s.handleGetStatus)
	mux.HandleFunc("GET /api/v1/status/{cluster}", s.handleGetStatusByCluster)
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /version", s.handleVersion)

	s.Handler = wrapMiddleware(mux,
		recoveryMiddleware,
		requestIDMiddleware,
		loggingMiddleware,
	)

	return s
}

// Start begins listening on the given address.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.Handler)
}
