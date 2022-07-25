package driver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kerraform/kegistry/internal/model"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	gOption "google.golang.org/api/option"
)

var (
	ErrGCSNotAllowed = errors.New("uploads and downloads are done by presigned url")
)

type GCS struct {
	bucket string
	logger *zap.Logger
	gcs    *storage.Client
}

type GCSOpts struct {
	Bucket            string
	ServiceAccountKey string
}

func newGCSDriver(ctx context.Context, logger *zap.Logger, opts *GCSOpts) (Driver, error) {
	gOptions := []gOption.ClientOption{}

	if key := opts.ServiceAccountKey; key != "" {
		gOptions = append(gOptions, gOption.WithCredentialsFile(key))
	}

	c, err := storage.NewClient(ctx, gOptions...)
	if err != nil {
		return nil, err
	}

	return &GCS{
		gcs: c,
	}, nil
}

func (d *GCS) CreateProvider(ctx context.Context, namespace, registryName string) error {
	return nil
}

func (d *GCS) CreateProviderPlatform(ctx context.Context, namespace, registryName, version, pos, arch string) (*CreateProviderPlatformResult, error) {
	binaryPath := fmt.Sprintf("%s/%s/%s/versions/%s/%s-%s/terraform-provider-%s_%s_%s_%s.zip", providerRootPath, namespace, registryName, version, pos, arch, registryName, version, pos, arch)
	psc := s3.NewPresignClient(d.s3)
	binaryUploadURL, err := psc.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(binaryPath),
	})
	if err != nil {
		return nil, err
	}

	return &CreateProviderPlatformResult{
		ProviderBinaryUploads: binaryUploadURL.URL,
	}, nil
}

func (d *GCS) CreateProviderVersion(ctx context.Context, namespace, registryName, version string) (*CreateProviderVersionResult, error) {
	versionRootPath := fmt.Sprintf("%s/%s/%s/versions/%s", providerRootPath, namespace, registryName, version)
	psc := s3.NewPresignClient(d.s3)
	sha256SumKeyUploadURL, err := psc.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS", versionRootPath, registryName, version)),
	})
	if err != nil {
		return nil, err
	}

	sha256SumSigKeyUploadURL, err := psc.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS.sig", versionRootPath, registryName, version)),
	})
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
	platformPath := fmt.Sprintf("%s/%s/%s/versions/%s/%s-%s", providerRootPath, namespace, registryName, version, pos, arch)
	filename := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", registryName, version, pos, arch)
	filepath := fmt.Sprintf("%s/%s", platformPath, filename)
	versionRootPath := fmt.Sprintf("%s/%s/%s/versions/%s", providerRootPath, namespace, registryName, version)
	keysRootPath := fmt.Sprintf("%s/%s/%s", providerRootPath, namespace, keyDirname)

	downloader := manager.NewDownloader(d.s3)
	wg, ctx := errgroup.WithContext(ctx)
	psc := s3.NewPresignClient(d.s3)
	var sha256Sum string
	var platformBinaryDownload *v4.PresignedHTTPRequest
	var sha256SumKeyDownload *v4.PresignedHTTPRequest
	var sha256SumSigKeyDownload *v4.PresignedHTTPRequest
	gpgKeys := []model.GPGPublicKey{}

	wg.Go(func() error {
		input := &s3.ListObjectsV2Input{
			Bucket: aws.String(d.bucket),
			Prefix: aws.String(keysRootPath),
		}

		resp, err := d.s3.ListObjectsV2(ctx, input)
		if err != nil {
			return err
		}

		d.logger.Debug("found keys",
			zap.Int("count", len(resp.Contents)),
			zap.String("path", keysRootPath),
		)

		for _, obj := range resp.Contents {
			b := manager.NewWriteAtBuffer([]byte{})
			_, err := downloader.Download(ctx, b, &s3.GetObjectInput{
				Bucket: aws.String(d.bucket),
				Key:    obj.Key,
			})
			if err != nil {
				return err
			}

			bs := bytes.NewBuffer(b.Bytes())
			block, err := armor.Decode(bs)
			if err != nil {
				return err
			}

			if block.Type != openpgp.PublicKeyType {
				return fmt.Errorf("not public key type")
			}

			reader := packet.NewReader(block.Body)
			pkt, err := reader.Next()
			if err != nil {
				return err
			}

			k, ok := pkt.(*packet.PublicKey)
			if !ok {
				return fmt.Errorf("not public key type")
			}

			gpgKey := model.GPGPublicKey{
				KeyID:      k.KeyIdString(),
				ASCIIArmor: string(b.Bytes()),
			}
			gpgKeys = append(gpgKeys, gpgKey)
		}

		return nil
	})

	wg.Go(func() error {
		b := manager.NewWriteAtBuffer([]byte{})
		_, err := downloader.Download(ctx, b, &s3.GetObjectInput{
			Bucket: aws.String(d.bucket),
			Key:    aws.String(filepath),
		})
		if err != nil {
			return err
		}

		sha256SumHex := sha256.Sum256(b.Bytes())
		sha256Sum = hex.EncodeToString(sha256SumHex[:])
		return nil
	})

	wg.Go(func() error {
		var err error
		platformBinaryDownload, err = psc.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(d.bucket),
			Key:    aws.String(filepath),
		})

		return err
	})

	wg.Go(func() error {
		var err error
		sha256SumKeyDownload, err = psc.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(d.bucket),
			Key:    aws.String(fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS", versionRootPath, registryName, version)),
		})

		return err
	})

	wg.Go(func() error {
		var err error
		sha256SumSigKeyDownload, err = psc.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(d.bucket),
			Key:    aws.String(fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS.sig", versionRootPath, registryName, version)),
		})

		return err
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	pkg := &model.Package{
		OS:            pos,
		Arch:          arch,
		Filename:      filename,
		DownloadURL:   platformBinaryDownload.URL,
		SHASumsURL:    sha256SumKeyDownload.URL,
		SHASumsSigURL: sha256SumSigKeyDownload.URL,
		SHASum:        sha256Sum,
		SigningKeys: &model.SigningKeys{
			GPGPublicKeys: gpgKeys,
		},
	}

	return pkg, nil
}

func (d *GCS) IsProviderCreated(ctx context.Context, namespace, registryName string) error {
	return nil
}

func (d *GCS) IsProviderVersionCreated(ctx context.Context, namespace, registryName, version string) error {
	return nil
}

func (d *GCS) ListAvailableVersions(ctx context.Context, namespace, registryName string) ([]model.AvailableVersion, error) {
	prefix := fmt.Sprintf("%s/%s/%s/versions", providerRootPath, namespace, registryName)
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

	platforms := map[string][]model.AvailableVersionPlatform{}
	for _, obj := range resp.Contents {
		ext := filepath.Ext(*obj.Key)
		if ext != ".zip" {
			continue
		}
		result := platformBinaryRegex.FindStringSubmatch(*obj.Key)

		if len(result) < 3 {
			continue
		}

		version := result[2]
		pos := result[3]
		arch := result[4]

		platforms[version] = append(platforms[version], model.AvailableVersionPlatform{
			OS:   pos,
			Arch: arch,
		})
	}

	vs := []model.AvailableVersion{}
	for k, versions := range platforms {
		vs = append(vs, model.AvailableVersion{
			Version:   k,
			Platforms: versions,
		})
	}

	return vs, nil
}

func (d *GCS) SaveGPGKey(ctx context.Context, namespace, keyID string, key []byte) error {
	keyPath := fmt.Sprintf("%s/%s/%s/%s", providerRootPath, namespace, keyDirname, keyID)
	w := d.Bucket().Object(keyPath).NewWriter(ctx)

	b := bytes.NewBuffer(key)
	if _, err := io.Copy(w, b); err != nil {
		return err
	}

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

	b := new(bytes.Buffer)
	metadata := &ProviderVersionMetadata{
		KeyID: keyID,
	}

	if err := json.NewEncoder(b).Encode(metadata); err != nil {
		return err
	}

	uploader := manager.NewUploader(d.s3)
	res, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(filepath),
		Body:   b,
	})
	if err != nil {
		return err
	}

	d.logger.Debug("save version metadata to amazon s3",
		zap.String("location", res.Location),
	)
	return nil
}

func (d *GCS) Bucket() *storage.BucketHandle {
	return d.gcs.Bucket(d.bucket)
}
