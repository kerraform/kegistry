package driver

import (
	"fmt"

	"go.uber.org/zap"
)

const (
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
	CreateProviderVersion(namespace, registryName, version string) error
	SaveGPGKey(namespace, key string) error
	IsProviderCreated(namespace, registryName string) error
	IsProviderVersionCreated(namespace, registryName, version string) error
}

type driverOpts struct {
	S3 *S3Opts
}

type DriverOpt func(opts *driverOpts)

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
