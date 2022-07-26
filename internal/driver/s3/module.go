package s3

import (
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
