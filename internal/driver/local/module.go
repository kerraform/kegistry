package local

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

type module struct {
	logger *zap.Logger
}

var _ driver.Module = (*module)(nil)

func (d *module) GetDownloadURL(ctx context.Context, namespace, provider, name, version string) (string, error) {
	return fmt.Sprintf("/registry/v1/modules/%s/%s/%s/%s/terraform-%s-%s-%v.tar.gz", namespace, provider, name, version, provider, name, version), nil
}

func (d *module) GetModule(ctx context.Context, namespace, provider, name, version string) (*os.File, error) {
	packagePath := fmt.Sprintf("%s/modules/%s/%s/%s/versions/%s/terraform-%s-%s-%v.tar.gz", localRootPath, namespace, provider, name, version, provider, name, version)
	return os.Open(packagePath)
}

func (d *module) ListAvailableVersions(ctx context.Context, namespace, provider, name string) ([]string, error) {
	modulePath := fmt.Sprintf("%s/%s/%s/%s/%s/versions", localRootPath, driver.ModuleRootPath, namespace, provider, name)
	fs, err := ioutil.ReadDir(modulePath)
	if err != nil {
		return nil, err
	}

	vs := []string{}
	for _, f := range fs {
		vs = append(vs, f.Name())
	}

	d.logger.Debug("list available versions",
		zap.Int("count", len(vs)),
		zap.String("path", modulePath),
	)
	return vs, nil
}
