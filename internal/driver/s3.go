package driver

import (
	"fmt"
	"io"

	"go.uber.org/zap"
	"golang.org/x/crypto/openpgp/packet"
)

type S3 struct {
	accessKey string
	logger    *zap.Logger
	secretKey string
}

type S3Opts struct {
	AccessKey string
	SecretKey string
}

func newS3Driver(logger *zap.Logger, opts *S3Opts) (Driver, error) {
	if opts == nil {
		return nil, fmt.Errorf("invalid s3 credentials")
	}
	return &S3{
		accessKey: opts.AccessKey,
		logger:    logger,
		secretKey: opts.SecretKey,
	}, nil
}

func (s3 *S3) CreateProvider(namespace, registryName string) error {
	return nil
}

func (s3 *S3) CreateProviderPlatform(namespace, registryName, version, pos, arch string) error {
	return nil
}

func (s3 *S3) CreateProviderVersion(namespace, registryName, version string) error {
	return nil
}

func (s3 *S3) IsProviderCreated(namespace, registryName string) error {
	return nil
}

func (s3 *S3) IsProviderVersionCreated(namespace, registryName, version string) error {
	return nil
}

func (s3 *S3) SaveGPGKey(namespace string, key *packet.PublicKey) error {
	return nil
}

func (s3 *S3) SavePlatformBinary(namespace, registryName, version, pos, arch string, body io.Reader) error {
	return nil
}

func (s3 *S3) SaveSHASUMs(namespace, registryName, version string, body io.Reader) error {
	return nil
}

func (s3 *S3) SaveSHASUMsSig(namespace, registryName, version string, body io.Reader) error {
	return nil
}
