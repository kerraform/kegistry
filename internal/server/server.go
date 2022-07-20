package server

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/middleware"
	v1 "github.com/kerraform/kegistry/internal/v1"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	baseURL *url.URL
	driver  driver.Driver
	logger  *zap.Logger
	mux     *mux.Router
	server  *http.Server

	v1 *v1.Handler
}

type ServerConfig struct {
	Driver driver.Driver
	Logger *zap.Logger

	V1 *v1.Handler
}

func NewServer(cfg *ServerConfig) *Server {
	s := &Server{
		driver: cfg.Driver,
		logger: cfg.Logger,
		mux:    mux.NewRouter(),
		v1:     cfg.V1,
	}

	s.registerRegistryHandler()
	s.registerUtilHandler()

	return s
}

func (s *Server) Serve(ctx context.Context, conn net.Listener) error {
	server := &http.Server{
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      middleware.Audit(s.logger)(s.mux),
	}

	s.server = server
	if err := server.Serve(conn); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
