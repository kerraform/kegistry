package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

type module struct {
	bucket string
	logger *zap.Logger
	s3     *s3.Client
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
