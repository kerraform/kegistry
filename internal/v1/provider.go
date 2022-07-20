package v1

import (
	"net/http"

	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"go.uber.org/zap"
)

type Provider interface {
	FindPackage() http.Handler
	ListAvailableVersions() http.Handler
}

type provider struct {
	driver driver.Driver
	logger *zap.Logger
}

var _ Provider = (*provider)(nil)

type providerConfig struct {
	Driver driver.Driver
	Logger *zap.Logger
}

func newProvider(cfg *providerConfig) Provider {
	return &provider{
		driver: cfg.Driver,
		logger: cfg.Logger,
	}
}

func (p *provider) FindPackage() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})
}

func (p *provider) ListAvailableVersions() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})
}
