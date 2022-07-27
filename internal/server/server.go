package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/metric"
	v1 "github.com/kerraform/kegistry/internal/v1"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	driver         *driver.Driver
	enableModule   bool
	enableProvider bool
	logger         *zap.Logger
	metric         *metric.RegistryMetrics
	mux            *mux.Router
	server         *http.Server

	v1 *v1.Handler
}

type ServerConfig struct {
	Driver         *driver.Driver
	EnableModule   bool
	EnableProvider bool
	Logger         *zap.Logger
	Metric         *metric.RegistryMetrics
	V1             *v1.Handler
}

func NewServer(cfg *ServerConfig) *Server {
	s := &Server{
		driver:         cfg.Driver,
		enableModule:   cfg.EnableModule,
		enableProvider: cfg.EnableProvider,
		logger:         cfg.Logger,
		metric:         cfg.Metric,
		mux:            mux.NewRouter(),
		v1:             cfg.V1,
	}

	s.metric.RegisterAllMetrics()

	s.registerRegistryHandler()
	s.registerUtilHandler()
	s.registerMetricsHandler()

	return s
}

func (s *Server) Serve(ctx context.Context, conn net.Listener) error {
	server := &http.Server{
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.mux,
	}

	s.metric.Resync(ctx)
	s.server = server
	if err := server.Serve(conn); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
