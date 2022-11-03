package local

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

type module struct {
	logger   *zap.Logger
	rootPath string
}

var _ driver.Module = (*module)(nil)

func (d *module) CreateModule(ctx context.Context, namespace, provider, name string) error {
	moduleRootPath := fmt.Sprintf("%s/modules/%s/%s/%s", d.rootPath, namespace, provider, name)
	if err := os.MkdirAll(moduleRootPath, 0700); err != nil {
		return err
	}
	d.logger.Debug("create module path", zap.String("path", moduleRootPath))
	return nil
}

func (d *module) CreateVersion(ctx context.Context, namespace, provider, name, version string) (*driver.CreateModuleVersionResult, error) {
	versionRootPath := fmt.Sprintf("%s/modules/%s/%s/%s/versions/%s", d.rootPath, namespace, provider, name, version)
	if err := os.MkdirAll(versionRootPath, 0700); err != nil {
		return nil, err
	}
	upload := fmt.Sprintf("/registry/v1/modules/%s/%s/%s/versions/%s", namespace, name, provider, version)
	d.logger.Debug("create module version path", zap.String("path", versionRootPath), zap.String("uploadPath", upload))
	return &driver.CreateModuleVersionResult{
		Upload: upload,
	}, nil
}

func (d *module) GetDownloadURL(ctx context.Context, namespace, provider, name, version string) (string, error) {
	return fmt.Sprintf("/registry/v1/modules/%s/%s/%s/%s/terraform-%s-%s-%v.tar.gz", namespace, provider, name, version, provider, name, version), nil
}

func (d *module) GetModule(ctx context.Context, namespace, provider, name, version string) (*os.File, error) {
	packagePath := fmt.Sprintf("%s/modules/%s/%s/%s/versions/%s/terraform-%s-%s-%s.tar.gz", d.rootPath, namespace, provider, name, version, provider, name, version)
	return os.Open(packagePath)
}

func (d *module) ListAvailableVersions(ctx context.Context, namespace, provider, name string) ([]string, error) {
	modulePath := fmt.Sprintf("%s/%s/%s/%s/%s/versions", d.rootPath, driver.ModuleRootPath, namespace, provider, name)
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

func (d *module) SavePackage(ctx context.Context, namespace, provider, name, version string, b io.Reader) error {
	pkgPath := fmt.Sprintf("%s/modules/%s/%s/%s/versions/%s/terraform-%s-%s-%s.tar.gz", d.rootPath, namespace, provider, name, version, provider, name, version)
	f, err := os.Create(pkgPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, b); err != nil {
		return err
	}
	defer f.Close()

	d.logger.Debug("create module version path", zap.String("path", pkgPath))
	return nil
}
