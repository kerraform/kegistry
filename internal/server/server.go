package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	driver driver.Driver
	logger *zap.Logger
	mux    *mux.Router
	server *http.Server
}

func NewServer(driver driver.Driver, logger *zap.Logger) *Server {
	s := &Server{
		driver: driver,
		logger: logger,
		mux:    mux.NewRouter(),
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
