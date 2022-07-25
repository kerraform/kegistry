package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/kerraform/kegistry/internal/model"
	"go.uber.org/zap"
)

var (
	ErrProviderBinaryNotExist  = errors.New("provider binary not exist")
	ErrProviderNotExist        = errors.New("provider not exist")
	ErrProviderVersionNotExist = errors.New("provider version not exist")

	platformBinaryRegex = regexp.MustCompile(`terraform-provider-(\w+)_([0-9]+.[0-9]+.[0-9]+)_(\w+)_(\w+).zip`)
)

const (
	keyDirname       = "keys"
	moduleRootPath   = "modules"
	providerRootPath = "providers"
)

type DriverType string

const (
	DriverTypeLocal DriverType = "local"
	DriverTypeGCS   DriverType = "gcs"
	DriverTypeS3    DriverType = "s3"
)

type Driver interface {
	CreateProvider(ctx context.Context, namespace, registryName string) error
	CreateProviderPlatform(ctx context.Context, namespace, registryName, version, os, arch string) (*CreateProviderPlatformResult, error)
	CreateProviderVersion(ctx context.Context, namespace, registryName, version string) (*CreateProviderVersionResult, error)
	GetPlatformBinary(ctx context.Context, namespace, registryName, version, os, arch string) (io.ReadCloser, error)
	GetSHASums(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error)
	GetSHASumsSig(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error)
	IsProviderCreated(ctx context.Context, namespace, registryName string) error
	IsProviderVersionCreated(ctx context.Context, namespace, registryName, version string) error
	SaveGPGKey(ctx context.Context, namespace, keyID string, key []byte) error
	SavePlatformBinary(ctx context.Context, namespace, registryName, version, os, arch string, body io.Reader) error
	SaveSHASUMs(ctx context.Context, namespace, registryName, version string, body io.Reader) error
	SaveSHASUMsSig(ctx context.Context, namespace, registryName, version string, body io.Reader) error
	SaveVersionMetadata(ctx context.Context, namespace, registryName, version, keyID string) error
	ListAvailableVersions(ctx context.Context, namespace, registryName string) ([]model.AvailableVersion, error)
	FindPackage(ctx context.Context, namespace, registryName, version, os, arch string) (*model.Package, error)
}

type driverOpts struct {
	GCS *GCSOpts
	S3  *S3Opts
}

type DriverOpt func(opts *driverOpts)

type ProviderVersionMetadata struct {
	KeyID string `json:"key-id"`
}

func WithGCS(gcsOpts *GCSOpts) DriverOpt {
	return func(opts *driverOpts) {
		opts.GCS = gcsOpts
	}
}

func WithS3(s3Opts *S3Opts) DriverOpt {
	return func(opts *driverOpts) {
		opts.S3 = s3Opts
	}
}

func NewDriver(ctx context.Context, driverType DriverType, logger *zap.Logger, opts ...DriverOpt) (Driver, error) {
	var o driverOpts
	for _, f := range opts {
		f(&o)
	}

	switch driverType {
	case DriverTypeLocal:
		return newLocalDriver(logger)
	case DriverTypeGCS:
		return newGCSDriver(ctx, logger, o.GCS)
	case DriverTypeS3:
		return newS3Driver(ctx, logger, o.S3)
	default:
		return nil, fmt.Errorf("no valid driver specified, got: %s", driverType)
	}
}

type CreateProviderVersionResult struct {
	SHASumsUpload    string
	SHASumsSigUpload string
}

type CreateProviderPlatformResult struct {
	ProviderBinaryUploads string
}
