package local

import (
	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

const (
	localRootPath = "/tmp"
)

func NewDriver(logger *zap.Logger) *driver.Driver {
	provider := &provider{
		logger: logger,
	}

	return &driver.Driver{
		Provider: provider,
	}
}
