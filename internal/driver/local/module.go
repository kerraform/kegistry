package local

import (
	"context"

	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

type module struct {
	logger *zap.Logger
}

var _ driver.Module = (*module)(nil)

func (d *module) GetDownloadURL(ctx context.Context, namespace, provider, name, version string) error {
	return nil
}

func (d *module) GetModule(ctx context.Context, namespace, provider, name, version string) error {
	return nil
}

func (d *module) ListAvailableVersions(ctx context.Context, namespace, provider, name string) error {
	return nil
}
