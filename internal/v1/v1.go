package v1

import (
	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

type Handler struct {
	Module   Module
	Provider Provider
}

type HandlerConfig struct {
	Driver driver.Driver
	Logger *zap.Logger
}

func New(cfg *HandlerConfig) *Handler {
	module := newModule(&moduleConfig{
		Driver: cfg.Driver,
		Logger: cfg.Logger,
	})

	provider := newProvider(&providerConfig{
		Driver: cfg.Driver,
		Logger: cfg.Logger,
	})

	return &Handler{
		Module:   module,
		Provider: provider,
	}
}
