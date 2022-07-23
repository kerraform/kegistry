package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"github.com/kerraform/kegistry/internal/model"
	"go.uber.org/zap"
)

type DataType string

const (
	DataTypeRegistryProviderVersions  DataType = "registry-provider-versions"
	DataTypeRegistryProviderPlatforms DataType = "registry-provider-platforms"
)

type Provider interface {
	CreateProvider() http.Handler
	CreateProviderPlatform() http.Handler
	CreateProviderVersion() http.Handler
	DownloadPlatformBinary() http.Handler
	DownloadSHASums() http.Handler
	DownloadSHASumsSignature() http.Handler
	FindPackage() http.Handler
	ListAvailableVersions() http.Handler
	UploadPlatformBinary() http.Handler
	UploadSHASums() http.Handler
	UploadSHASumsSignature() http.Handler
}

type provider struct {
	driver driver.Driver
	logger *zap.Logger
}

var _ Provider = (*provider)(nil)

type providerConfig struct {
	Driver driver.Driver
	Logger *zap.Logger
}

func newProvider(cfg *providerConfig) Provider {
	return &provider{
		driver: cfg.Driver,
		logger: cfg.Logger,
	}
}

//
// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#request-body
type CreateProviderRequest struct {
	Data *CreateProviderRequestData `json:"data"`
}

type CreateProviderRequestData struct {
	Type       DataType                             `json:"type"`
	Attributes *CreateProviderRequestDataAttributes `json:"attributes"`
}

type CreateProviderRequestDataAttributes struct {
	// Name of the provider (e.g. aws)
	Name string `json:"name"`

	// Name of the namespace (a.k.a. organization)
	Namespace string `json:"namespace"`
}

func (p *provider) CreateProvider() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		var req CreateProviderRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}
		defer r.Body.Close()

		if err := p.driver.CreateProvider(req.Data.Attributes.Namespace, req.Data.Attributes.Name); err != nil {
			return err
		}
		p.logger.Info("provisioned provider path", zap.String("name", req.Data.Attributes.Name), zap.String("namespace", req.Data.Attributes.Namespace))

		w.WriteHeader(http.StatusOK)
		return nil
	})
}

type CreateProviderPlatformRequest struct {
	Data *CreateProviderPlatformRequestData `json:"data"`
}

type CreateProviderPlatformRequestData struct {
	Type       DataType                                     `json:"type"`
	Attributes *CreateProviderPlatformRequestDataAttributes `json:"attributes"`
}

type CreateProviderPlatformRequestDataAttributes struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type CreateProviderPlatformResponse struct {
	Data *CreateProviderPlatformResponseData `json:"data"`
}

type CreateProviderPlatformResponseData struct {
	Type  DataType                                `json:"type"`
	Links *CreateProviderPlatformResponseDataLink `json:"attributes"`
}

type CreateProviderPlatformResponseDataLink struct {
	ProviderBinaryUploads string `json:"provider-binary-upload"`
}

func (p *provider) CreateProviderPlatform() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		var req CreateProviderPlatformRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
			zap.String("version", version),
			zap.String("os", req.Data.Attributes.OS),
			zap.String("arch", req.Data.Attributes.Arch),
		)

		if err := p.driver.CreateProviderPlatform(namespace, registryName, version, req.Data.Attributes.OS, req.Data.Attributes.Arch); err != nil {
			return err
		}
		defer r.Body.Close()

		l.Info("provisioned provider platform")

		resp := &CreateProviderPlatformResponse{
			Data: &CreateProviderPlatformResponseData{
				Type: DataTypeRegistryProviderPlatforms,
				Links: &CreateProviderPlatformResponseDataLink{
					ProviderBinaryUploads: fmt.Sprintf("/v1/providers/%s/%s/versions/%s/%s/%s/binary", namespace, registryName, version, req.Data.Attributes.OS, req.Data.Attributes.Arch),
				},
			},
		}
		return json.NewEncoder(w).Encode(resp)
	})
}

//
// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#request-body
type CreateProviderVersionRequest struct {
	Data *CreateProviderVersionRequestData `json:"data"`
}

type CreateProviderVersionRequestData struct {
	Type       DataType                                    `json:"type"`
	Attributes *CreateProviderVersionRequestDataAttributes `json:"attributes"`
}

type CreateProviderVersionRequestDataAttributes struct {
	// Version of the provider in semver (e.g. v2.0.1)
	Version string `json:"version"`

	// Valid gpg-key string
	KeyID string `json:"key-id"`
}

type CreateProviderVersionResponse struct {
	Data *CreateProviderVersionResponseData `json:"data"`
}

type CreateProviderVersionResponseData struct {
	Type  DataType                               `json:"type"`
	Links *CreateProviderVersionResponseDataLink `json:"attributes"`
}

type CreateProviderVersionResponseDataLink struct {
	SHASumsUpload    string `json:"shasums-upload"`
	SHASumsSigUpload string `json:"shasums-sig-upload"`
}

func (p *provider) CreateProviderVersion() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]

		var req CreateProviderVersionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}
		defer r.Body.Close()

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
		)

		if err := p.driver.IsProviderCreated(namespace, registryName); err != nil {
			if errors.Is(err, driver.ErrProviderNotExist) {
				l.Error("not provider found")
				w.WriteHeader(http.StatusBadRequest)
				return err
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		if err := p.driver.CreateProviderVersion(namespace, registryName, req.Data.Attributes.Version); err != nil {
			l.Error("failed to create provider version")
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		if err := p.driver.SaveVersionMetadata(namespace, registryName, req.Data.Attributes.Version, req.Data.Attributes.KeyID); err != nil {
			l.Error("failed to save provider version metadata")
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		l.Info("create provider version")
		w.WriteHeader(http.StatusOK)

		resp := &CreateProviderVersionResponse{
			Data: &CreateProviderVersionResponseData{
				Type: DataTypeRegistryProviderVersions,
				Links: &CreateProviderVersionResponseDataLink{
					SHASumsUpload:    fmt.Sprintf("/v1/providers/%s/%s/versions/%s/sigsums", namespace, registryName, req.Data.Attributes.Version),
					SHASumsSigUpload: fmt.Sprintf("/v1/providers/%s/%s/versions/%s/shasums-sig", namespace, registryName, req.Data.Attributes.Version),
				},
			},
		}

		return json.NewEncoder(w).Encode(resp)
	})
}

func (p *provider) DownloadPlatformBinary() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]
		os := mux.Vars(r)["os"]
		arch := mux.Vars(r)["arch"]

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
			zap.String("version", version),
			zap.String("os", os),
			zap.String("arch", arch),
		)

		if err := p.driver.GetPlatformBinary(namespace, registryName, version, os, arch); err != nil {
			return err
		}
		defer r.Body.Close()

		l.Info("saved platform binary")
		return nil
	})
}

func (p *provider) DownloadSHASums() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
			zap.String("version", version),
		)

		if err := p.driver.GetSHASums(namespace, registryName, version); err != nil {
			return err
		}
		defer r.Body.Close()

		l.Info("saved platform binary")
		return nil
	})
}

func (p *provider) DownloadSHASumsSignature() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
			zap.String("version", version),
		)

		if err := p.driver.GetSHASumsSig(namespace, registryName, version); err != nil {
			return err
		}
		defer r.Body.Close()

		l.Info("saved platform binary")
		return nil
	})
}

func (p *provider) FindPackage() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]
		os := mux.Vars(r)["os"]
		arch := mux.Vars(r)["arch"]

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
			zap.String("version", version),
			zap.String("os", os),
			zap.String("arch", arch),
		)

		if err := p.driver.IsProviderCreated(namespace, registryName); err != nil {
			if errors.Is(err, driver.ErrProviderNotExist) {
				l.Error("not provider found")
				w.WriteHeader(http.StatusBadRequest)
				return err
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		if err := p.driver.IsProviderVersionCreated(namespace, registryName, version); err != nil {
			if errors.Is(err, driver.ErrProviderNotExist) {
				l.Error("not provider version found")
				w.WriteHeader(http.StatusBadRequest)
				return err
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		pkg, err := p.driver.FindPackage(namespace, registryName, version, os, arch)
		if err != nil {
			if errors.Is(err, driver.ErrProviderBinaryNotExist) {
				w.WriteHeader(http.StatusNotFound)
				return err
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		l.Info("package found")

		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(pkg)
	})
}

type ListAvailableVersionsResponse struct {
	Versions []model.AvailableVersion `json:"versions"`
}

type AvailableVersion struct {
	Version   string                           `json:"version"`
	Protocols []string                         `json:"protocols"`
	Platforms []model.AvailableVersionPlatform `json:"platforms"`
}

func (p *provider) ListAvailableVersions() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]

		versions, err := p.driver.ListAvailableVersions(namespace, registryName)
		if err != nil {
			return err
		}

		resp := &ListAvailableVersionsResponse{
			Versions: versions,
		}

		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(resp)
	})
}

func (p *provider) UploadPlatformBinary() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]
		os := mux.Vars(r)["os"]
		arch := mux.Vars(r)["arch"]

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
			zap.String("version", version),
			zap.String("os", os),
			zap.String("arch", arch),
		)

		if err := p.driver.SavePlatformBinary(namespace, registryName, version, os, arch, r.Body); err != nil {
			return err
		}
		defer r.Body.Close()

		l.Info("saved platform binary")
		return nil
	})
}

func (p *provider) UploadSHASums() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
			zap.String("version", version),
		)

		if err := p.driver.SaveSHASUMs(namespace, registryName, version, r.Body); err != nil {
			return err
		}
		defer r.Body.Close()

		l.Info("saved shasums")
		return nil
	})
}

func (p *provider) UploadSHASumsSignature() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		l := p.logger.With(
			zap.String("namespace", namespace),
			zap.String("registryName", registryName),
			zap.String("version", version),
		)

		if err := p.driver.SaveSHASUMsSig(namespace, registryName, version, r.Body); err != nil {
			return err
		}
		defer r.Body.Close()

		l.Info("saved shasums signature")
		return nil
	})
}
