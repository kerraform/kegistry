package s3

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kerraform/kegistry/internal/driver"
	"go.uber.org/zap"
)

type module struct {
	bucket string
	logger *zap.Logger
	s3     *s3.Client
}

var _ driver.Module = (*module)(nil)

func (d *module) CreateModule(ctx context.Context, namespace, provider, name string) error {
	return nil
}

func (d *module) CreateVersion(ctx context.Context, namespace, provider, name, version string) (*driver.CreateModuleVersionResult, error) {
	filepath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/terraform-%s-%s-%s.tar.gz", driver.ModuleRootPath, namespace, provider, name, version, provider, name, version)
	psc := s3.NewPresignClient(d.s3)
	uploadURL, err := psc.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(filepath),
	})
	if err != nil {
		return nil, err
	}

	return &driver.CreateModuleVersionResult{
		Upload: uploadURL.URL,
	}, nil
}

func (d *module) GetDownloadURL(ctx context.Context, namespace, provider, name, version string) (string, error) {
	filepath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/terraform-%s-%s-%s.tar.gz", driver.ModuleRootPath, namespace, provider, name, version, provider, name, version)

	psc := s3.NewPresignClient(d.s3)
	downloadURL, err := psc.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(filepath),
	})

	if err != nil {
		return "", err
	}

	return downloadURL.URL, nil
}

func (d *module) GetModule(ctx context.Context, namespace, provider, name, version string) (*os.File, error) {
	return nil, ErrS3NotAllowed
}

func (d *module) ListAvailableVersions(ctx context.Context, namespace, provider, name string) ([]string, error) {
	prefix := fmt.Sprintf("%s/%s/%s/%s/versions", driver.ModuleRootPath, namespace, provider, name)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(d.bucket),
		Prefix: aws.String(prefix),
	}

	resp, err := d.s3.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}

	d.logger.Debug("found versions",
		zap.Int("count", len(resp.Contents)),
	)

	vs := []string{}
	for _, obj := range resp.Contents {
		vs = append(vs, *obj.Key)
	}

	return vs, nil
}

func (d *module) SavePackage(ctx context.Context, namespace, provider, name, version string, body io.Reader) error {
	return ErrS3NotAllowed
}
