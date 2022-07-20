package driver

import (
	"fmt"

	"go.uber.org/zap"
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

func (s3 *S3) CreateProviderPlatform(namespace, registryName string) error {
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
