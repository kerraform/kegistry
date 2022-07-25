package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"cloud.google.com/go/storage"
	"github.com/kerraform/kegistry/internal/model"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	gOption "google.golang.org/api/option"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

var (
	ErrGCSNotAllowed = errors.New("uploads and downloads are done by presigned url")
)

type GCS struct {
	bucket string
	logger *zap.Logger
	iam    *credentials.IamCredentialsClient
	gcs    *storage.Client
}

type GCSOpts struct {
	Bucket string
}

func newGCSDriver(ctx context.Context, logger *zap.Logger, opts *GCSOpts) (Driver, error) {
	gOptions := []gOption.ClientOption{}
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}

	gOptions = append(gOptions, gOption.WithCredentials(creds))
	gcsClient, err := storage.NewClient(ctx, gOptions...)
	if err != nil {
		return nil, err
	}

	iamClient, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCS{
		bucket: opts.Bucket,
		gcs:    gcsClient,
		iam:    iamClient,
		logger: logger.With(zap.String("bucket", opts.Bucket)),
	}, nil
}

func (d *GCS) CreateProvider(ctx context.Context, namespace, registryName string) error {
	return nil
}

func (d *GCS) CreateProviderPlatform(ctx context.Context, namespace, registryName, version, pos, arch string) (*CreateProviderPlatformResult, error) {
	binaryPath := fmt.Sprintf("%s/%s/%s/versions/%s/%s-%s/terraform-provider-%s_%s_%s_%s.zip", providerRootPath, namespace, registryName, version, pos, arch, registryName, version, pos, arch)
	opts := &storage.SignedURLOptions{
		Expires: time.Now().Add(15 * time.Minute),
		Method:  http.MethodGet,
		Scheme:  storage.SigningSchemeV4,
	}

	opts.SignBytes = func(b []byte) ([]byte, error) {
		req := &credentialspb.SignBlobRequest{
			Payload: b,
		}

		resp, err := d.iam.SignBlob(ctx, req)
		if err != nil {
			return nil, err
		}

		return resp.SignedBlob, err
	}

	u, err := storage.SignedURL(d.bucket, fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS", binaryPath, registryName, version), opts)
	if err != nil {
		return nil, err
	}

	return &CreateProviderPlatformResult{
		ProviderBinaryUploads: u,
	}, nil
}

func (d *GCS) CreateProviderVersion(ctx context.Context, namespace, registryName, version string) (*CreateProviderVersionResult, error) {
	versionRootPath := fmt.Sprintf("%s/%s/%s/versions/%s", providerRootPath, namespace, registryName, version)
	opts := &storage.PostPolicyV4Options{
		// Scheme:  storage.SigningSchemeV4,
		Expires: time.Now().Add(15 * time.Minute),
		// Method:  http.MethodGet,
	}

	opts.SignRawBytes = func(b []byte) ([]byte, error) {
		req := &credentialspb.SignBlobRequest{
			Payload: b,
		}

		resp, err := d.iam.SignBlob(ctx, req)
		if err != nil {
			return nil, err
		}

		return resp.SignedBlob, err
	}

	sha256SumKeyUploadURL, err := d.Bucket().GenerateSignedPostPolicyV4(fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS", versionRootPath, registryName, version), opts)
	if err != nil {
		return nil, err
	}

	sha256SumSigKeyUploadURL, err := d.Bucket().GenerateSignedPostPolicyV4(fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS.sig", versionRootPath, registryName, version), opts)
	if err != nil {
		return nil, err
	}

	d.logger.Debug("created provider version path", zap.String("path", versionRootPath))
	return &CreateProviderVersionResult{
		SHASumsUpload:    sha256SumKeyUploadURL.URL,
		SHASumsSigUpload: sha256SumSigKeyUploadURL.URL,
	}, nil
}

func (d *GCS) GetPlatformBinary(ctx context.Context, namespace, registryName, version, pos, arch string) (io.ReadCloser, error) {
	return nil, ErrGCSNotAllowed
}

func (d *GCS) GetSHASums(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error) {
	return nil, ErrGCSNotAllowed
}

func (d *GCS) GetSHASumsSig(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error) {
	return nil, ErrGCSNotAllowed
}

func (d *GCS) FindPackage(ctx context.Context, namespace, registryName, version, pos, arch string) (*model.Package, error) {
	return nil, nil
}

func (d *GCS) IsProviderCreated(ctx context.Context, namespace, registryName string) error {
	return nil
}

func (d *GCS) IsProviderVersionCreated(ctx context.Context, namespace, registryName, version string) error {
	return nil
}

func (d *GCS) ListAvailableVersions(ctx context.Context, namespace, registryName string) ([]model.AvailableVersion, error) {
	prefix := fmt.Sprintf("%s/%s/%s/versions", providerRootPath, namespace, registryName)

	it := d.Bucket().Objects(ctx, &storage.Query{
		Prefix: prefix,
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		fmt.Println(attrs.Name)
	}

	return nil, nil
}

func (d *GCS) SaveGPGKey(ctx context.Context, namespace, keyID string, key []byte) error {
	keyPath := fmt.Sprintf("%s/%s/%s/%s", providerRootPath, namespace, keyDirname, keyID)
	w := d.Bucket().Object(keyPath).NewWriter(ctx)
	b := bytes.NewBuffer(key)
	if _, err := io.Copy(w, b); err != nil {
		return err
	}
	defer w.Close()

	d.logger.Debug("saved gpg key to gcs", zap.String("path", keyPath))
	return nil
}

func (d *GCS) SavePlatformBinary(ctx context.Context, namespace, registryName, version, pos, arch string, body io.Reader) error {
	return ErrGCSNotAllowed
}

func (d *GCS) SaveSHASUMs(ctx context.Context, namespace, registryName, version string, body io.Reader) error {
	return ErrGCSNotAllowed
}

func (d *GCS) SaveSHASUMsSig(ctx context.Context, namespace, registryName, version string, body io.Reader) error {
	return ErrGCSNotAllowed
}

func (d *GCS) SaveVersionMetadata(ctx context.Context, namespace, registryName, version, keyID string) error {
	filepath := fmt.Sprintf("%s/%s/%s/versions/%s/%s", providerRootPath, namespace, registryName, version, versionMetadataFilename)
	if err := d.IsProviderVersionCreated(ctx, namespace, registryName, version); err != nil {
		return err
	}

	w := d.Bucket().Object(filepath).NewWriter(ctx)
	metadata := &ProviderVersionMetadata{
		KeyID: keyID,
	}

	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		return err
	}

	d.logger.Debug("save version metadata to gcs",
		zap.String("path", filepath),
	)
	return nil
}

func (d *GCS) Bucket() *storage.BucketHandle {
	return d.gcs.Bucket(d.bucket)
}
