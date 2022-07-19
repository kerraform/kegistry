package server

import (
	"net/http"
)

func (s *Server) registerUtilHandler() {
	s.mux.Methods(http.MethodGet).Path("/healthz").Handler(s.HealthCheck())
}

func (s *Server) HealthCheck() http.Handler {
	return handler(s.logger, func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})
}
