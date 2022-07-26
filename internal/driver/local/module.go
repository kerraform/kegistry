package local

import (
	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

type module struct {
	logger *zap.Logger
}

var _ driver.Module = (*module)(nil)
