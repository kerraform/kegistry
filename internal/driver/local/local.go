package local

import (
	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

const (
	localRootPath = "/tmp"
)

func NewDriver(logger *zap.Logger) *driver.Driver {
	module := &module{
		logger: logger,
	}

	provider := &provider{
		logger: logger,
	}

	return &driver.Driver{
		Module:   module,
		Provider: provider,
	}
}
