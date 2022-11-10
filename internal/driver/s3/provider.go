package s3

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

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/kerraform/kegistry/internal/driver"
	model "github.com/kerraform/kegistry/internal/model/provider"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type provider struct {
	bucket string
	logger *zap.Logger
	s3     *s3.Client
	tracer trace.Tracer
}

type endpointResolver struct {
	URL string
}

func (r *endpointResolver) ResolveEndpoint(service, region string, options ...interface{}) (aws.Endpoint, error) {
	return aws.Endpoint{
		URL: r.URL,
	}, nil
}

func (d *provider) CreateProvider(ctx context.Context, namespace, registryName string) error {
	return nil
}

func (d *provider) CreateProviderPlatform(ctx context.Context, namespace, registryName, version, pos, arch string) (*driver.CreateProviderPlatformResult, error) {
	ctx, span := d.tracer.Start(ctx, "CreateProviderPlatform")
	defer span.End()
	binaryPath := fmt.Sprintf("%s/%s/%s/versions/%s/%s-%s/terraform-provider-%s_%s_%s_%s.zip", driver.ProviderRootPath, namespace, registryName, version, pos, arch, registryName, version, pos, arch)
	psc := s3.NewPresignClient(d.s3)
	binaryUploadURL, err := psc.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(binaryPath),
	})
	if err != nil {
		return nil, err
	}

	return &driver.CreateProviderPlatformResult{
		ProviderBinaryUploads: binaryUploadURL.URL,
	}, nil
}

func (d *provider) CreateProviderVersion(ctx context.Context, namespace, registryName, version string) (*driver.CreateProviderVersionResult, error) {
	ctx, span := d.tracer.Start(ctx, "CreateProviderVersion")
	defer span.End()
	versionRootPath := fmt.Sprintf("%s/%s/%s/versions/%s", driver.ProviderRootPath, namespace, registryName, version)
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

	return &driver.CreateProviderVersionResult{
		SHASumsUpload:    sha256SumKeyUploadURL.URL,
		SHASumsSigUpload: sha256SumSigKeyUploadURL.URL,
	}, nil
}

func (d *provider) GetPlatformBinary(ctx context.Context, namespace, registryName, version, pos, arch string) (io.ReadCloser, error) {
	return nil, ErrS3NotAllowed
}

func (d *provider) GetSHASums(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error) {
	return nil, ErrS3NotAllowed
}

func (d *provider) GetSHASumsSig(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error) {
	return nil, ErrS3NotAllowed
}

func (d *provider) FindPackage(ctx context.Context, namespace, registryName, version, pos, arch string) (*model.Package, error) {
	ctx, span := d.tracer.Start(ctx, "FindPackage")
	defer span.End()

	platformPath := fmt.Sprintf("%s/%s/%s/versions/%s/%s-%s", driver.ProviderRootPath, namespace, registryName, version, pos, arch)
	filename := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", registryName, version, pos, arch)
	filepath := fmt.Sprintf("%s/%s", platformPath, filename)
	versionRootPath := fmt.Sprintf("%s/%s/%s/versions/%s", driver.ProviderRootPath, namespace, registryName, version)
	keysRootPath := fmt.Sprintf("%s/%s/%s", driver.ProviderRootPath, namespace, driver.KeyDirname)

	downloader := manager.NewDownloader(d.s3)
	wg, ctx := errgroup.WithContext(ctx)
	psc := s3.NewPresignClient(d.s3)
	var sha256Sum string
	var platformBinaryDownload *v4.PresignedHTTPRequest
	var sha256SumKeyDownload *v4.PresignedHTTPRequest
	var sha256SumSigKeyDownload *v4.PresignedHTTPRequest
	gpgKeys := []model.GPGPublicKey{}

	wg.Go(func() error {
		newCtx, span := d.tracer.Start(ctx, "gpgKeys")
		defer span.End()
		input := &s3.ListObjectsV2Input{
			Bucket: aws.String(d.bucket),
			Prefix: aws.String(keysRootPath),
		}

		resp, err := d.s3.ListObjectsV2(newCtx, input)
		if err != nil {
			return err
		}

		d.logger.Debug("found keys",
			zap.Int("count", len(resp.Contents)),
			zap.String("path", keysRootPath),
		)

		for _, obj := range resp.Contents {
			newCtx2, span2 := d.tracer.Start(newCtx, "gpgKey")
			b := manager.NewWriteAtBuffer([]byte{})
			_, err := downloader.Download(newCtx2, b, &s3.GetObjectInput{
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
			span2.End()
		}

		return nil
	})

	wg.Go(func() error {
		newCtx, span := d.tracer.Start(ctx, "sha256sum")
		defer span.End()
		b := manager.NewWriteAtBuffer([]byte{})
		_, err := downloader.Download(newCtx, b, &s3.GetObjectInput{
			Bucket: aws.String(d.bucket),
			Key:    aws.String(filepath),
		})

		if err != nil {
			var ae smithy.APIError
			if errors.As(err, &ae) {
				return driver.ErrProviderBinaryNotExist
			}

			return err
		}

		sha256SumHex := sha256.Sum256(b.Bytes())
		sha256Sum = hex.EncodeToString(sha256SumHex[:])
		return nil
	})

	wg.Go(func() error {
		newCtx, span := d.tracer.Start(ctx, "downloadURL")
		defer span.End()
		var err error
		platformBinaryDownload, err = psc.PresignGetObject(newCtx, &s3.GetObjectInput{
			Bucket: aws.String(d.bucket),
			Key:    aws.String(filepath),
		})

		if err != nil {
			var ae smithy.APIError
			if errors.As(err, &ae) {
				return driver.ErrProviderBinaryNotExist
			}

			return err
		}

		return nil
	})

	wg.Go(func() error {
		newCtx, span := d.tracer.Start(ctx, "sha256SumsURL")
		defer span.End()
		var err error
		sha256SumKeyDownload, err = psc.PresignGetObject(newCtx, &s3.GetObjectInput{
			Bucket: aws.String(d.bucket),
			Key:    aws.String(fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS", versionRootPath, registryName, version)),
		})

		if err != nil {
			var ae smithy.APIError
			if errors.As(err, &ae) {
				return driver.ErrProviderSHA256SUMSNotExist
			}

			return err
		}

		return err
	})

	wg.Go(func() error {
		newCtx, span := d.tracer.Start(ctx, "sha256SumSigURL")
		defer span.End()
		var err error
		sha256SumSigKeyDownload, err = psc.PresignGetObject(newCtx, &s3.GetObjectInput{
			Bucket: aws.String(d.bucket),
			Key:    aws.String(fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS.sig", versionRootPath, registryName, version)),
		})

		if err != nil {
			var ae smithy.APIError
			if errors.As(err, &ae) {
				return driver.ErrProviderSHA256SUMSSigNotExist
			}

			return err
		}

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

func (d *provider) IsGPGKeyCreated(ctx context.Context, namespace, registryName string) error {
	ctx, span := d.tracer.Start(ctx, "IsGPGKeyCreated")
	defer span.End()
	keyPath := fmt.Sprintf("%s/%s/%s", driver.ProviderRootPath, namespace, driver.KeyDirname)
	objs, err := d.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(d.bucket),
		Delimiter: aws.String(keyPath),
	})

	if err != nil {
		var bne *types.NotFound
		if errors.As(err, &bne) {
			return driver.ErrProviderGPGKeyNotExist
		}

		return err
	}

	if len(objs.Contents) == 0 {
		return driver.ErrProviderGPGKeyNotExist
	}

	return nil
}

func (d *provider) IsProviderCreated(ctx context.Context, namespace, registryName string) error {
	return nil
}

func (d *provider) IsProviderVersionCreated(ctx context.Context, namespace, registryName, version string) error {
	return nil
}

func (d *provider) ListAvailableVersions(ctx context.Context, namespace, registryName string) ([]model.AvailableVersion, error) {
	ctx, span := d.tracer.Start(ctx, "ListAvailableVersions")
	defer span.End()
	prefix := fmt.Sprintf("%s/%s/%s/versions", driver.ProviderRootPath, namespace, registryName)
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
		result := driver.PlatformBinaryRegex.FindStringSubmatch(*obj.Key)

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

func (d *provider) SaveGPGKey(ctx context.Context, namespace, keyID string, key []byte) error {
	ctx, span := d.tracer.Start(ctx, "SaveGPGKey")
	defer span.End()
	keyPath := fmt.Sprintf("%s/%s/%s/%s", driver.ProviderRootPath, namespace, driver.KeyDirname, keyID)
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

func (d *provider) SavePlatformBinary(ctx context.Context, namespace, registryName, version, pos, arch string, body io.Reader) error {
	// Note: There are no "directory" system on Amazon S3
	return ErrS3NotAllowed
}

func (d *provider) SaveSHASUMs(ctx context.Context, namespace, registryName, version string, body io.Reader) error {
	// Note: There are no "directory" system on Amazon S3
	return ErrS3NotAllowed
}

func (d *provider) SaveSHASUMsSig(ctx context.Context, namespace, registryName, version string, body io.Reader) error {
	// Note: There are no "directory" system on Amazon S3
	return ErrS3NotAllowed
}

func (d *provider) SaveVersionMetadata(ctx context.Context, namespace, registryName, version, keyID string) error {
	ctx, span := d.tracer.Start(ctx, "SaveVersionMetadata")
	defer span.End()
	filepath := fmt.Sprintf("%s/%s/%s/versions/%s/%s", driver.ProviderRootPath, namespace, registryName, version, driver.VersionMetadataFilename)
	if err := d.IsProviderVersionCreated(ctx, namespace, registryName, version); err != nil {
		return err
	}

	b := new(bytes.Buffer)
	metadata := &driver.ProviderVersionMetadata{
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
