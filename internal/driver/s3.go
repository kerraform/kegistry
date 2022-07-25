package driver

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kerraform/kegistry/internal/model"
	"go.uber.org/zap"
)

type S3 struct {
	bucket string
	logger *zap.Logger
	s3     *s3.Client
}

type S3Opts struct {
	AccessKey string
	Bucket    string
	Endpoint  string
	SecretKey string
}

type endpointResolver struct {
	URL string
}

func (r *endpointResolver) ResolveEndpoint(service, region string, options ...interface{}) (aws.Endpoint, error) {
	return aws.Endpoint{
		URL: r.URL,
	}, nil
}

func newS3Driver(logger *zap.Logger, opts *S3Opts) (Driver, error) {
	if opts == nil {
		return nil, fmt.Errorf("invalid s3 credentials")
	}

	endpointResolver := &endpointResolver{
		URL: opts.Endpoint,
	}

	cred := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.SecretKey, ""))
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(cred),
		config.WithEndpointResolverWithOptions(endpointResolver),
	)
	if err != nil {
		return nil, err
	}

	return &S3{
		bucket: opts.Bucket,
		logger: logger,
		s3:     s3.NewFromConfig(cfg),
	}, nil
}

func (d *S3) CreateProvider(ctx context.Context, namespace, registryName string) error {
	return nil
}

func (d *S3) CreateProviderPlatform(ctx context.Context, namespace, registryName, version, osName, arch string) error {
	return nil
}

func (d *S3) CreateProviderVersion(ctx context.Context, namespace, registryName, version string) error {
	return nil
}

func (d *S3) GetPlatformBinary(ctx context.Context, namespace, registryName, version, pos, arch string) (io.ReadCloser, error) {
	return nil, nil
}

func (d *S3) GetSHASums(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error) {
	return nil, nil
}

func (d *S3) GetSHASumsSig(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error) {
	return nil, nil
}

func (d *S3) FindPackage(ctx context.Context, namespace, registryName, version, os, arch string) (*model.Package, error) {
	return nil, nil
}

func (d *S3) IsProviderCreated(ctx context.Context, namespace, registryName string) error {
	return nil
}

func (d *S3) IsProviderVersionCreated(ctx context.Context, namespace, registryName, version string) error {
	return nil
}

func (d *S3) SaveGPGKey(ctx context.Context, namespace, keyID string, key []byte) error {
	keyPath := fmt.Sprintf("%s/%s/%s/%s", providerRootPath, namespace, keyDirname, keyID)
	uploader := manager.NewUploader(d.s3)

	b := bytes.NewBuffer(key)
	res, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(keyPath),
		Body:   b,
	})
	if err != nil {
		return err
	}

	d.logger.Debug("saved gpg key to amazon s3", zap.String("location", res.Location))
	return nil
}

func (d *S3) SavePlatformBinary(ctx context.Context, namespace, registryName, version, pos, arch string, body io.Reader) error {
	return nil
}

func (d *S3) SaveSHASUMs(ctx context.Context, namespace, registryName, version string, body io.Reader) error {
	return nil
}

func (d *S3) SaveSHASUMsSig(ctx context.Context, namespace, registryName, version string, body io.Reader) error {
	return nil
}

func (d *S3) SaveVersionMetadata(ctx context.Context, namespace, registryName, version, keyID string) error {
	return nil
}

func (d *S3) ListAvailableVersions(ctx context.Context, namespace, registryName string) ([]model.AvailableVersion, error) {
	return []model.AvailableVersion{}, nil
}
