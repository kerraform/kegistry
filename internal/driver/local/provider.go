package local

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kerraform/kegistry/internal/driver"
	model "github.com/kerraform/kegistry/internal/model/provider"
	"go.uber.org/zap"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

type provider struct {
	logger *zap.Logger
}

var _ driver.Provider = (*provider)(nil)

func (d *provider) CreateProvider(ctx context.Context, namespace, registryName string) error {
	registryRootPath := fmt.Sprintf("%s/%s/%s/%s", localRootPath, driver.ProviderRootPath, namespace, registryName)
	if err := os.MkdirAll(registryRootPath, 0700); err != nil {
		return err
	}
	d.logger.Debug("created registry path", zap.String("path", registryRootPath))
	return nil
}

func (d *provider) CreateProviderPlatform(ctx context.Context, namespace, registryName, version, pos, arch string) (*driver.CreateProviderPlatformResult, error) {
	platformRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s-%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version, pos, arch)
	if err := os.MkdirAll(platformRootPath, 0700); err != nil {
		return nil, err
	}
	d.logger.Debug("created platform path", zap.String("path", platformRootPath))
	return &driver.CreateProviderPlatformResult{
		ProviderBinaryUploads: fmt.Sprintf("/registry/v1/providers/%s/%s/versions/%s/%s/%s/binary", namespace, registryName, version, pos, arch),
	}, nil
}

func (d *provider) CreateProviderVersion(ctx context.Context, namespace, registryName, version string) (*driver.CreateProviderVersionResult, error) {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version)
	if err := os.MkdirAll(versionRootPath, 0700); err != nil {
		return nil, err
	}
	d.logger.Debug("created version path", zap.String("path", versionRootPath))
	return &driver.CreateProviderVersionResult{
		SHASumsUpload:    fmt.Sprintf("/registry/v1/providers/%s/%s/versions/%s/shasums", namespace, registryName, version),
		SHASumsSigUpload: fmt.Sprintf("/registry/v1/providers/%s/%s/versions/%s/shasums-sig", namespace, registryName, version),
	}, nil
}

func (d *provider) GetPlatformBinary(ctx context.Context, namespace, registryName, version, pos, arch string) (io.ReadCloser, error) {
	platformPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s-%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version, pos, arch)
	filename := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", registryName, version, pos, arch)
	filepath := fmt.Sprintf("%s/%s", platformPath, filename)
	return os.Open(filepath)
}

func (d *provider) GetSHASums(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error) {
	filepath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/terraform-provider-%s_%s_SHA256SUMS", localRootPath, driver.ProviderRootPath, namespace, registryName, version, registryName, version)
	return os.Open(filepath)
}

func (d *provider) GetSHASumsSig(ctx context.Context, namespace, registryName, version string) (io.ReadCloser, error) {
	filepath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/terraform-provider-%s_%s_SHA256SUMS.sig", localRootPath, driver.ProviderRootPath, namespace, registryName, version, registryName, version)
	return os.Open(filepath)
}

func (d *provider) FindPackage(ctx context.Context, namespace, registryName, version, pos, arch string) (*model.Package, error) {
	platformPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s-%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version, pos, arch)
	filename := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", registryName, version, pos, arch)
	filepath := fmt.Sprintf("%s/%s", platformPath, filename)

	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			d.logger.Error("file not exist", zap.String("filepath", filepath))
			return nil, driver.ErrProviderBinaryNotExist
		}

		return nil, err
	}

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	sha256SumHex := sha256.Sum256(b)
	sha256Sum := hex.EncodeToString(sha256SumHex[:])
	d.logger.Debug("generated sha256sum for file", zap.String("sha256sum", sha256Sum))

	keysPath := fmt.Sprintf("%s/%s/%s/%s", localRootPath, driver.ProviderRootPath, namespace, driver.KeyDirname)
	keys, err := ioutil.ReadDir(keysPath)
	if err != nil {
		return nil, err
	}

	gpgKeys := []model.GPGPublicKey{}
	for _, key := range keys {
		d.logger.Debug("found keys",
			zap.Int("count", len(keys)),
			zap.String("path", key.Name()),
		)
		f, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", keysPath, key.Name()))
		if err != nil {
			d.logger.Error("failed to read key file",
				zap.String("file", key.Name()),
				zap.Error(err),
			)

			continue
		}

		b := bytes.NewBuffer(f)
		block, err := armor.Decode(b)
		if err != nil {
			d.logger.Error("failed to armor decode", zap.Error(err))
			continue
		}

		if block.Type != openpgp.PublicKeyType {
			d.logger.Error("skipping non public key type")
			continue
		}

		reader := packet.NewReader(block.Body)
		pkt, err := reader.Next()
		if err != nil {
			d.logger.Error("failed to packet read", zap.Error(err))
			continue
		}

		k, ok := pkt.(*packet.PublicKey)
		if !ok {
			d.logger.Error("failed to cast packet to public key type")
			continue
		}

		gpgKey := model.GPGPublicKey{
			KeyID:      k.KeyIdString(),
			ASCIIArmor: string(f),
		}
		gpgKeys = append(gpgKeys, gpgKey)
	}

	signingKeys := &model.SigningKeys{
		GPGPublicKeys: gpgKeys,
	}

	pkg := &model.Package{
		OS:            pos,
		Arch:          arch,
		Filename:      filename,
		DownloadURL:   fmt.Sprintf("/registry/v1/providers/%s/%s/versions/%s/%s/%s/binary", namespace, registryName, version, pos, arch),
		SHASumsURL:    fmt.Sprintf("/registry/v1/providers/%s/%s/versions/%s/shasums", namespace, registryName, version),
		SHASumsSigURL: fmt.Sprintf("/registry/v1/providers/%s/%s/versions/%s/shasums-sig", namespace, registryName, version),
		SHASum:        sha256Sum,
		SigningKeys:   signingKeys,
	}

	return pkg, nil
}

func (d *provider) IsProviderCreated(ctx context.Context, namespace, registryName string) error {
	registryRootPath := fmt.Sprintf("%s/%s/%s/%s", localRootPath, driver.ProviderRootPath, namespace, registryName)
	d.logger.Debug("checking provider", zap.String("path", registryRootPath))
	if _, err := os.Stat(registryRootPath); err != nil {
		if os.IsNotExist(err) {
			return driver.ErrProviderNotExist
		}

		return err
	}

	return nil
}

func (d *provider) IsProviderVersionCreated(ctx context.Context, namespace, registryName, version string) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version)
	d.logger.Debug("checking provider version", zap.String("path", versionRootPath))
	if _, err := os.Stat(versionRootPath); err != nil {
		if os.IsNotExist(err) {
			return driver.ErrProviderNotExist
		}

		return err
	}

	return nil
}

func (d *provider) ListAvailableVersions(ctx context.Context, namespace, registryName string) ([]model.AvailableVersion, error) {
	versionsRootPath := fmt.Sprintf("%s/%s/%s/%s/versions", localRootPath, driver.ProviderRootPath, namespace, registryName)
	versions, err := ioutil.ReadDir(versionsRootPath)
	if err != nil {
		return nil, err
	}

	d.logger.Debug("found versions", zap.String("path", versionsRootPath), zap.Int("count", len(versions)))

	vs := make([]model.AvailableVersion, len(versions))
	for i, version := range versions {
		if !version.IsDir() {
			d.logger.Debug("skip file in version directory", zap.String("version", version.Name()))
			continue
		}
		//
		platforms, err := ioutil.ReadDir(filepath.Join(versionsRootPath, version.Name()))
		if err != nil {
			return nil, err
		}

		d.logger.Debug("found platforms for this version",
			zap.Int("count", len(platforms)),
			zap.String("version", version.Name()),
		)

		pfs := []model.AvailableVersionPlatform{}

		for _, platform := range platforms {
			if !platform.IsDir() {
				d.logger.Debug("skip file in platform directory", zap.String("platform", platform.Name()))
				continue
			}

			e := strings.Split(platform.Name(), "-")

			pfs = append(pfs, model.AvailableVersionPlatform{
				OS:   e[0],
				Arch: e[1],
			})
		}

		vs[i] = model.AvailableVersion{
			Version:   version.Name(),
			Platforms: pfs,
		}
	}

	return vs, nil
}

func (d *provider) SaveGPGKey(ctx context.Context, namespace, keyID string, key []byte) error {
	keyRootPath := fmt.Sprintf("%s/%s/%s/%s", localRootPath, driver.ProviderRootPath, namespace, driver.KeyDirname)
	if err := os.MkdirAll(keyRootPath, 0700); err != nil {
		return err
	}

	keyPath := fmt.Sprintf("%s/%s", keyRootPath, keyID)
	f, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(key); err != nil {
		return err
	}
	d.logger.Debug("saved gpg key", zap.String("filepath", keyPath))
	return nil
}

func (d *provider) SavePlatformBinary(ctx context.Context, namespace, registryName, version, pos, arch string, body io.Reader) error {
	platformPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s-%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version, pos, arch)
	if err := d.IsProviderVersionCreated(ctx, namespace, registryName, version); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/terraform-provider-%s_%s_%s_%s.zip", platformPath, registryName, version, pos, arch)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, body)
	d.logger.Debug("save platform binary",
		zap.String("path", filepath),
	)
	return err
}

func (d *provider) SaveSHASUMs(ctx context.Context, namespace, registryName, version string, body io.Reader) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version)
	if err := d.IsProviderVersionCreated(ctx, namespace, registryName, version); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS", versionRootPath, registryName, version)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, body)
	d.logger.Debug("save shasums",
		zap.String("path", filepath),
	)
	return err
}

func (d *provider) SaveSHASUMsSig(ctx context.Context, namespace, registryName, version string, body io.Reader) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version)
	if err := d.IsProviderVersionCreated(ctx, namespace, registryName, version); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS.sig", versionRootPath, registryName, version)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, body)
	d.logger.Debug("save shasums signature",
		zap.String("path", filepath),
	)
	return err
}

func (d *provider) SaveVersionMetadata(ctx context.Context, namespace, registryName, version, keyID string) error {
	filepath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s", localRootPath, driver.ProviderRootPath, namespace, registryName, version, driver.VersionMetadataFilename)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	b := new(bytes.Buffer)
	metadata := &driver.ProviderVersionMetadata{
		KeyID: keyID,
	}

	if err := json.NewEncoder(b).Encode(metadata); err != nil {
		return err
	}

	_, err = io.Copy(f, b)
	d.logger.Debug("save version metadata",
		zap.String("path", filepath),
	)
	return err
}