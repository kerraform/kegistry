package s3

import (
	"context"
	"io"
	"os"

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

func (d *module) CreateModule(ctx context.Context, namespace, provider, name string) error {
	return nil
}

func (d *module) CreateVersion(ctx context.Context, namespace, provider, name, version string) (*driver.CreateModuleVersionResult, error) {
	return nil, nil
}

func (d *module) GetDownloadURL(ctx context.Context, namespace, provider, name, version string) (string, error) {
	return "", nil
}

func (d *module) GetModule(ctx context.Context, namespace, provider, name, version string) (*os.File, error) {
	return nil, nil
}

func (d *module) ListAvailableVersions(ctx context.Context, namespace, provider, name string) ([]string, error) {
	return nil, nil
}

func (d *module) SavePackage(ctx context.Context, namespace, provider, name, version string, body io.Reader) error {
	return nil
}
