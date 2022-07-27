package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

var (
	ErrS3NotAllowed = errors.New("uploads to s3 are done by presigned url")
)

type DriverOpts struct {
	AccessKey    string
	Bucket       string
	Endpoint     string
	SecretKey    string
	UsePathStyle bool
}

func NewDriver(logger *zap.Logger, opts *DriverOpts) (*driver.Driver, error) {
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

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = opts.UsePathStyle
	})

	module := &module{
		bucket: opts.Bucket,
		logger: logger,
		s3:     s3Client,
	}

	provider := &provider{
		bucket: opts.Bucket,
		logger: logger,
		s3:     s3Client,
	}

	return &driver.Driver{
		Module:   module,
		Provider: provider,
	}, nil
}
