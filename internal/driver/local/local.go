package local

import (
	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

type DriverConfig struct {
	RootPath string
	Logger   *zap.Logger
}

func NewDriver(cfg *DriverConfig) *driver.Driver {
	module := &module{
		rootPath: cfg.RootPath,
		logger:   cfg.Logger,
	}

	provider := &provider{
		rootPath: cfg.RootPath,
		logger:   cfg.Logger,
	}

	return &driver.Driver{
		Module:   module,
		Provider: provider,
	}
}
