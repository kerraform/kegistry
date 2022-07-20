package driver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kerraform/kegistry/internal/model"
	"go.uber.org/zap"
)

var (
	ErrProviderNotExist        = errors.New("provider not exist")
	ErrProviderVersionNotExist = errors.New("provider version not exist")
)

const (
	localRootPath = "/tmp"
)

type local struct {
	logger *zap.Logger
}

var _ Driver = (*local)(nil)

func newLocalDriver(logger *zap.Logger) (Driver, error) {
	return &local{
		logger: logger,
	}, nil
}

func (l *local) CreateProvider(namespace, registryName string) error {
	registryRootPath := fmt.Sprintf("%s/%s/%s/%s", localRootPath, providerRootPath, namespace, registryName)
	if err := os.MkdirAll(registryRootPath, 0700); err != nil {
		return err
	}
	l.logger.Debug("created registry path", zap.String("path", registryRootPath))
	return nil
}

func (l *local) CreateProviderPlatform(namespace, registryName, version, osName, arch, filename string) error {
	packageRootDir := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s-%s", localRootPath, providerRootPath, namespace, registryName, version, osName, arch)
	if err := os.MkdirAll(packageRootDir, 0700); err != nil {
		return err
	}
	l.logger.Debug("created version path", zap.String("path", packageRootDir))
	return nil
}

func (l *local) CreateProviderVersion(namespace, registryName, version string) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, providerRootPath, namespace, registryName, version)
	if err := os.MkdirAll(versionRootPath, 0700); err != nil {
		return err
	}
	l.logger.Debug("created version path", zap.String("path", versionRootPath))
	return nil
}

func (l *local) FindPackage(namespace, registryName, version, os, arch string) (*model.Package, error) {
	return nil, nil
}

func (l *local) IsProviderCreated(namespace, registryName string) error {
	registryRootPath := fmt.Sprintf("%s/%s/%s/%s", localRootPath, providerRootPath, namespace, registryName)
	l.logger.Debug("checking provider", zap.String("path", registryRootPath))
	if _, err := os.Stat(registryRootPath); err != nil {
		if os.IsNotExist(err) {
			return ErrProviderNotExist
		}

		return err
	}

	return nil
}

func (l *local) IsProviderVersionCreated(namespace, registryName, version string) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, providerRootPath, namespace, registryName, version)
	l.logger.Debug("checking provider version", zap.String("path", versionRootPath))
	if _, err := os.Stat(versionRootPath); err != nil {
		if os.IsNotExist(err) {
			return ErrProviderNotExist
		}

		return err
	}

	return nil
}

func (l *local) ListAvailableVersions(namespace, registryName string) ([]model.AvailableVersion, error) {
	versionsRootPath := fmt.Sprintf("%s/%s/%s/%s/versions", localRootPath, providerRootPath, namespace, registryName)
	versions, err := ioutil.ReadDir(versionsRootPath)
	if err != nil {
		return nil, err
	}

	l.logger.Debug("found versions", zap.String("path", versionsRootPath), zap.Int("count", len(versions)))

	vs := make([]model.AvailableVersion, len(versions))
	for i, version := range versions {
		platforms, err := ioutil.ReadDir(filepath.Join(versionsRootPath, version.Name()))
		if err != nil {
			return nil, err
		}

		pfs := make([]model.AvailableVersionPlatform, len(platforms))

		for j, platform := range platforms {
			e := strings.Split(platform.Name(), "-")

			pfs[j] = model.AvailableVersionPlatform{
				OS:   e[0],
				Arch: e[1],
			}
		}

		vs[i] = model.AvailableVersion{
			Version:   version.Name(),
			Platforms: pfs,
		}
	}

	return vs, nil
}
