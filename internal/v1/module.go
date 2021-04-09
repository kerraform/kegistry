package v1

import (
	"net/http"

	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"go.uber.org/zap"
)

type Module interface {
	Download() http.Handler
	ListAvailableVersions() http.Handler
}

type module struct {
	driver driver.Driver
	logger *zap.Logger
}

var _ Module = (*module)(nil)

type moduleConfig struct {
	Driver driver.Driver
	Logger *zap.Logger
}

func newModule(cfg *moduleConfig) Module {
	return &module{
		driver: cfg.Driver,
		logger: cfg.Logger,
	}
}

func (m *module) Download() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})
}

func (m *module) ListAvailableVersions() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})
}
