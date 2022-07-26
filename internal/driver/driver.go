package driver

import (
	"context"
	"errors"
	"io"
	"regexp"

	"github.com/kerraform/kegistry/internal/model/provider"
)

var (
	ErrProviderBinaryNotExist  = errors.New("provider binary not exist")
	ErrProviderNotExist        = errors.New("provider not exist")
	ErrProviderVersionNotExist = errors.New("provider version not exist")

	PlatformBinaryRegex = regexp.MustCompile(`terraform-provider-(\w+)_([0-9]+.[0-9]+.[0-9]+)_(\w+)_(\w+).zip`)
)

const (
	KeyDirname              = "keys"
	ModuleRootPath          = "modules"
	ProviderRootPath        = "providers"
	VersionMetadataFilename = "metadata.json"
)

type DriverType string

const (
	DriverTypeLocal DriverType = "local"
	DriverTypeS3    DriverType = "s3"
)

type Module interface {
	GetDownloadURL(ctx context.Context, namespace, provider, name, version string) error
	GetModule(ctx context.Context, namespace, provider, name, version string) error
	ListAvailableVersions(ctx context.Context, namespace, provider, name string) error
}

type Provider interface {
	CreateProvider(ctx context.Context, namespace, registryName string) error
	CreateProviderPlatform(ctx context.Context, namespace, registryName, version, os, arch string) (*CreateProviderPlatformResult, error)
	CreateProviderVersion(ctx context.Context, namespace, registryName, version string) (*CreateProviderVersionResult, error)
	FindPackage(ctx context.Context, namespace, registryName, version, os, arch string) (*provider.Package, error)
	GetPlatformBinary(ctx context.Context, namespace, registryName, version, os, arch string) (io.ReadCloser, error)
	GetSHASums(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error)
	GetSHASumsSig(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error)
	ListAvailableVersions(ctx context.Context, namespace, registryName string) ([]provider.AvailableVersion, error)
	IsProviderCreated(ctx context.Context, namespace, registryName string) error
	IsProviderVersionCreated(ctx context.Context, namespace, registryName, version string) error
	SaveGPGKey(ctx context.Context, namespace, keyID string, key []byte) error
	SavePlatformBinary(ctx context.Context, namespace, registryName, version, os, arch string, body io.Reader) error
	SaveSHASUMs(ctx context.Context, namespace, registryName, version string, body io.Reader) error
	SaveSHASUMsSig(ctx context.Context, namespace, registryName, version string, body io.Reader) error
	SaveVersionMetadata(ctx context.Context, namespace, registryName, version, keyID string) error
}

type Driver struct {
	Module   Module
	Provider Provider
}

type ProviderVersionMetadata struct {
	KeyID string `json:"key-id"`
}

type CreateProviderVersionResult struct {
	SHASumsUpload    string
	SHASumsSigUpload string
}

type CreateProviderPlatformResult struct {
	ProviderBinaryUploads string
}
