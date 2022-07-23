package driver

import (
	"errors"
	"fmt"
	"io"

	"github.com/kerraform/kegistry/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/openpgp/packet"
)

var (
	ErrProviderBinaryNotExist  = errors.New("provider binary not exist")
	ErrProviderNotExist        = errors.New("provider not exist")
	ErrProviderVersionNotExist = errors.New("provider version not exist")
)

const (
	keyDirname       = "keys"
	moduleRootPath   = "modules"
	providerRootPath = "providers"
)

type DriverType string

const (
	DriverTypeLocal DriverType = "local"
	DriverTypeS3    DriverType = "s3"
)

type Driver interface {
	CreateProvider(namespace, registryName string) error
	CreateProviderPlatform(namespace, registryName, version, os, arch string) error
	CreateProviderVersion(namespace, registryName, version string) error
	GetPlatformBinary(namespace, registryName, version, os, arch string) (io.ReadCloser, error)
	GetSHASums(namespace, registryName, version string) (io.ReadCloser, error)
	GetSHASumsSig(namespace, registryName, version string) (io.ReadCloser, error)
	IsProviderCreated(namespace, registryName string) error
	IsProviderVersionCreated(namespace, registryName, version string) error
	SaveGPGKey(namespace string, key *packet.PublicKey) error
	SavePlatformBinary(namespace, registryName, version, os, arch string, body io.Reader) error
	SaveSHASUMs(namespace, registryName, version string, body io.Reader) error
	SaveSHASUMsSig(namespace, registryName, version string, body io.Reader) error
	SaveVersionMetadata(namespace, registryName, version, keyID string) error
	ListAvailableVersions(namespace, registryName string) ([]model.AvailableVersion, error)
	FindPackage(namespace, registryName, version, os, arch string) (*model.Package, error)
}

type driverOpts struct {
	S3 *S3Opts
}

type DriverOpt func(opts *driverOpts)

type ProviderVersionMetadata struct {
	KeyID string `json:"key-id"`
}

func WithS3(s3Opts *S3Opts) DriverOpt {
	return func(opts *driverOpts) {
		opts.S3 = s3Opts
	}
}

func NewDriver(driverType DriverType, logger *zap.Logger, opts ...DriverOpt) (Driver, error) {
	var o driverOpts
	for _, f := range opts {
		f(&o)
	}

	switch driverType {
	case DriverTypeS3:
		return newS3Driver(logger, o.S3)
	case DriverTypeLocal:
		return newLocalDriver(logger)
	default:
		return nil, fmt.Errorf("no valid driver specified, got: %s", driverType)
	}
}
