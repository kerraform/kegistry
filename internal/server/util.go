package server

import (
	"errors"
	"net/http"

	"github.com/kerraform/kegistry/internal/handler"
	"github.com/kerraform/kegistry/internal/middleware"
)

func (s *Server) registerUtilHandler() {
	s.mux.Methods(http.MethodGet).Path("/healthz").Handler(s.HealthCheck())

	s.mux.NotFoundHandler = middleware.AccessLog(s.logger)(s.NotFound())
}

func (s *Server) HealthCheck() http.Handler {
	return handler.NewHandler(func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})
}

func (s *Server) NotFound() http.Handler {
	return handler.NewHandler(func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		return errors.New("not found")
	})
}
