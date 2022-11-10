package local

import (
	"github.com/kerraform/kegistry/internal/driver"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type DriverConfig struct {
	RootPath string
	Logger   *zap.Logger
	Tracer   trace.Tracer
}

func NewDriver(cfg *DriverConfig) *driver.Driver {
	module := &module{
		logger:   cfg.Logger,
		rootPath: cfg.RootPath,
		tracer:   cfg.Tracer,
	}

	provider := &provider{
		logger:   cfg.Logger,
		rootPath: cfg.RootPath,
		tracer:   cfg.Tracer,
	}

	return &driver.Driver{
		Module:   module,
		Provider: provider,
	}
}
