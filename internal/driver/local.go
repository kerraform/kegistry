package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kerraform/kegistry/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

const (
	localRootPath           = "/tmp"
	versionMetadataFilename = "metadata.json"
)

type local struct {
	logger *zap.Logger
}

var _ Driver = (*local)(nil)

func newLocalDriver(logger *zap.Logger) (Driver, error) {
	return &local{
		logger: logger,
	}, nil
}

func (l *local) CreateProvider(namespace, registryName string) error {
	registryRootPath := fmt.Sprintf("%s/%s/%s/%s", localRootPath, providerRootPath, namespace, registryName)
	if err := os.MkdirAll(registryRootPath, 0700); err != nil {
		return err
	}
	l.logger.Debug("created registry path", zap.String("path", registryRootPath))
	return nil
}

func (l *local) CreateProviderPlatform(namespace, registryName, version, pos, arch string) error {
	platformRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s-%s", localRootPath, providerRootPath, namespace, registryName, version, pos, arch)
	if err := os.MkdirAll(platformRootPath, 0700); err != nil {
		return err
	}
	l.logger.Debug("created platform path", zap.String("path", platformRootPath))
	return nil
}

func (l *local) CreateProviderVersion(namespace, registryName, version string) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, providerRootPath, namespace, registryName, version)
	if err := os.MkdirAll(versionRootPath, 0700); err != nil {
		return err
	}
	l.logger.Debug("created version path", zap.String("path", versionRootPath))
	return nil
}

func (l *local) FindPackage(namespace, registryName, version, pos, arch string) (*model.Package, error) {
	platformPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s-%s", localRootPath, providerRootPath, namespace, registryName, version, pos, arch)
	filepath := fmt.Sprintf("%s/terraform-provider-%s_%s_%s_%s.zip", platformPath, registryName, version, pos, arch)

	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProviderBinaryNotExist
		}

		return nil, err
	}

	pkg := &model.Package{
		OS:            pos,
		Arch:          arch,
		DownloadURL:   "",
		SHASumsURL:    "",
		SHASumsSigURL: "",
		SHASum:        "",
	}

	return pkg, nil
}

func (l *local) IsProviderCreated(namespace, registryName string) error {
	registryRootPath := fmt.Sprintf("%s/%s/%s/%s", localRootPath, providerRootPath, namespace, registryName)
	l.logger.Debug("checking provider", zap.String("path", registryRootPath))
	if _, err := os.Stat(registryRootPath); err != nil {
		if os.IsNotExist(err) {
			return ErrProviderNotExist
		}

		return err
	}

	return nil
}

func (l *local) IsProviderVersionCreated(namespace, registryName, version string) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, providerRootPath, namespace, registryName, version)
	l.logger.Debug("checking provider version", zap.String("path", versionRootPath))
	if _, err := os.Stat(versionRootPath); err != nil {
		if os.IsNotExist(err) {
			return ErrProviderNotExist
		}

		return err
	}

	return nil
}

func (l *local) ListAvailableVersions(namespace, registryName string) ([]model.AvailableVersion, error) {
	versionsRootPath := fmt.Sprintf("%s/%s/%s/%s/versions", localRootPath, providerRootPath, namespace, registryName)
	versions, err := ioutil.ReadDir(versionsRootPath)
	if err != nil {
		return nil, err
	}

	l.logger.Debug("found versions", zap.String("path", versionsRootPath), zap.Int("count", len(versions)))

	vs := make([]model.AvailableVersion, len(versions))
	for i, version := range versions {
		if !version.IsDir() {
			l.logger.Debug("skip file in version directory", zap.String("version", version.Name()))
			continue
		}
		//
		platforms, err := ioutil.ReadDir(filepath.Join(versionsRootPath, version.Name()))
		if err != nil {
			return nil, err
		}

		l.logger.Debug("found platforms for this version",
			zap.Int("count", len(platforms)),
			zap.String("version", version.Name()),
		)

		pfs := []model.AvailableVersionPlatform{}

		for _, platform := range platforms {
			if !platform.IsDir() {
				l.logger.Debug("skip file in platform directory", zap.String("platform", platform.Name()))
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

func (local *local) SaveGPGKey(namespace string, key *packet.PublicKey) error {
	keyRootPath := fmt.Sprintf("%s/%s/%s", localRootPath, providerRootPath, namespace)
	if err := os.MkdirAll(keyRootPath, 0700); err != nil {
		return err
	}

	keyPath := fmt.Sprintf("%s/%s", keyRootPath, key.KeyIdString())
	f, err := os.Create(keyPath)
	if err != nil {
		return err
	}

	w, err := armor.Encode(f, openpgp.PublicKeyType, make(map[string]string))
	if err != nil {
		return err
	}

	if err := key.Serialize(w); err != nil {
		return err
	}
	defer w.Close()

	return nil
}

func (l *local) SavePlatformBinary(namespace, registryName, version, pos, arch string, body io.Reader) error {
	platformPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s-%s", localRootPath, providerRootPath, namespace, registryName, version, pos, arch)
	if err := l.IsProviderVersionCreated(namespace, registryName, version); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/terraform-provider-%s_%s_%s_%s.zip", platformPath, registryName, version, pos, arch)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, body)
	l.logger.Debug("save platform binary",
		zap.String("path", filepath),
	)
	return err
}

func (l *local) SaveSHASUMs(namespace, registryName, version string, body io.Reader) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, providerRootPath, namespace, registryName, version)
	if err := l.IsProviderVersionCreated(namespace, registryName, version); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS", versionRootPath, registryName, version)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, body)
	l.logger.Debug("save shasums",
		zap.String("path", filepath),
	)
	return err
}

func (l *local) SaveSHASUMsSig(namespace, registryName, version string, body io.Reader) error {
	versionRootPath := fmt.Sprintf("%s/%s/%s/%s/versions/%s", localRootPath, providerRootPath, namespace, registryName, version)
	if err := l.IsProviderVersionCreated(namespace, registryName, version); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/terraform-provider-%s_%s_SHA256SUMS.sig", versionRootPath, registryName, version)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, body)
	l.logger.Debug("save shasums signature",
		zap.String("path", filepath),
	)
	return err
}

func (l *local) SaveVersionMetadata(namespace, registryName, version, keyID string) error {
	filepath := fmt.Sprintf("%s/%s/%s/%s/versions/%s/%s", localRootPath, providerRootPath, namespace, registryName, version, versionMetadataFilename)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	b := new(bytes.Buffer)
	metadata := &ProviderVersionMetadata{
		KeyID: keyID,
	}

	if err := json.NewEncoder(b).Encode(metadata); err != nil {
		return err
	}

	_, err = io.Copy(f, b)
	return err
}
