package provider

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"go.uber.org/zap"
)

type DataType string

const (
	DataTypeRegistryProviderVersions  DataType = "registry-provider-versions"
	DataTypeRegistryProviderPlatforms DataType = "registry-provider-platforms"
)

type Provider struct {
	driver *driver.Driver
	logger *zap.Logger
}

type Config struct {
	Driver *driver.Driver
	Logger *zap.Logger
}

func New(cfg *Config) *Provider {
	return &Provider{
		driver: cfg.Driver,
		logger: cfg.Logger,
	}
}

// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#request-body
type CreateProviderRequestDataAttributes struct {
	// Name of the provider (e.g. aws)
	Name string `json:"name"`

	// Name of the namespace (a.k.a. organization)
	Namespace string `json:"namespace"`
}

func (p *Provider) CreateProvider() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		var req CreateProviderRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}
		defer r.Body.Close()

		if err := p.driver.Provider.CreateProvider(r.Context(), req.Data.Attributes.Namespace, req.Data.Attributes.Name); err != nil {
			return err
		}
		p.logger.Info("provisioned provider path", zap.String("name", req.Data.Attributes.Name), zap.String("namespace", req.Data.Attributes.Namespace))

		w.WriteHeader(http.StatusOK)
		return nil
	})
}

type CreateProviderPlatformRequestDataAttributes struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type CreateProviderPlatformResponseData struct {
	Links *CreateProviderPlatformResponseDataLink `json:"attributes"`
	Type  DataType                                `json:"type"`
}

type CreateProviderPlatformResponseDataLink struct {
	ProviderBinaryUploads string `json:"provider-binary-upload"`
}

func (p *Provider) CreateProviderPlatform() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		var req CreateProviderPlatformRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		result, err := p.driver.Provider.CreateProviderPlatform(r.Context(), namespace, registryName, version, req.Data.Attributes.OS, req.Data.Attributes.Arch)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		resp := &CreateProviderPlatformResponse{
			Data: &CreateProviderPlatformResponseData{
				Type: DataTypeRegistryProviderPlatforms,
				Links: &CreateProviderPlatformResponseDataLink{
					ProviderBinaryUploads: result.ProviderBinaryUploads,
				},
			},
		}
		return json.NewEncoder(w).Encode(resp)
	})
}

// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#request-body

type CreateProviderVersionRequestDataAttributes struct {
	// Version of the provider in semver (e.g. v2.0.1)
	Version string `json:"version"`

	// Valid gpg-key string
	KeyID string `json:"key-id"`
}

type CreateProviderVersionResponseData struct {
	Links *CreateProviderVersionResponseDataLink `json:"attributes"`
	Type  DataType                               `json:"type"`
}

type CreateProviderVersionResponseDataLink struct {
	SHASumsUpload    string `json:"shasums-upload"`
	SHASumsSigUpload string `json:"shasums-sig-upload"`
}

func (p *Provider) CreateProviderVersion() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]

		var req CreateProviderVersionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}
		defer r.Body.Close()

		if err := p.driver.Provider.IsProviderCreated(r.Context(), namespace, registryName); err != nil {
			if errors.Is(err, driver.ErrProviderNotExist) {
				p.logger.Error("not provider found")
				w.WriteHeader(http.StatusBadRequest)
				return err
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		result, err := p.driver.Provider.CreateProviderVersion(r.Context(), namespace, registryName, req.Data.Attributes.Version)
		if err != nil {
			p.logger.Error("failed to create provider version")
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		if err := p.driver.Provider.SaveVersionMetadata(r.Context(), namespace, registryName, req.Data.Attributes.Version, req.Data.Attributes.KeyID); err != nil {
			p.logger.Error("failed to save provider version metadata")
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		w.WriteHeader(http.StatusOK)

		resp := &CreateProviderVersionResponse{
			Data: &CreateProviderVersionResponseData{
				Type: DataTypeRegistryProviderVersions,
				Links: &CreateProviderVersionResponseDataLink{
					SHASumsUpload:    result.SHASumsUpload,
					SHASumsSigUpload: result.SHASumsSigUpload,
				},
			},
		}

		return json.NewEncoder(w).Encode(resp)
	})
}

func (p *Provider) DownloadPlatformBinary() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]
		os := mux.Vars(r)["os"]
		arch := mux.Vars(r)["arch"]

		f, err := p.driver.Provider.GetPlatformBinary(r.Context(), namespace, registryName, version, os, arch)
		if err != nil {
			return err
		}
		defer f.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := io.Copy(w, f); err != nil {
			return err
		}

		return nil
	})
}

func (p *Provider) DownloadSHASums() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		f, err := p.driver.Provider.GetSHASums(r.Context(), namespace, registryName, version)
		if err != nil {
			return err
		}
		defer f.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := io.Copy(w, f); err != nil {
			return err
		}

		return nil
	})
}

func (p *Provider) DownloadSHASumsSignature() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		f, err := p.driver.Provider.GetSHASumsSig(r.Context(), namespace, registryName, version)
		if err != nil {
			return err
		}
		defer f.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := io.Copy(w, f); err != nil {
			return err
		}

		return nil
	})
}

func (p *Provider) FindPackage() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]
		os := mux.Vars(r)["os"]
		arch := mux.Vars(r)["arch"]

		if err := p.driver.Provider.IsProviderCreated(r.Context(), namespace, registryName); err != nil {
			if errors.Is(err, driver.ErrProviderNotExist) {
				p.logger.Error("not provider found")
				w.WriteHeader(http.StatusBadRequest)
				return err
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		if err := p.driver.Provider.IsProviderVersionCreated(r.Context(), namespace, registryName, version); err != nil {
			if errors.Is(err, driver.ErrProviderNotExist) {
				l.Error("not provider version found")
				w.WriteHeader(http.StatusBadRequest)
				return err
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		pkg, err := p.driver.Provider.FindPackage(r.Context(), namespace, registryName, version, os, arch)
		if err != nil {
			if errors.Is(err, driver.ErrProviderBinaryNotExist) {
				w.WriteHeader(http.StatusNotFound)
				return err
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(pkg)
	})
}

func (p *Provider) ListAvailableVersions() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]

		versions, err := p.driver.Provider.ListAvailableVersions(r.Context(), namespace, registryName)
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

func (p *Provider) UploadPlatformBinary() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]
		os := mux.Vars(r)["os"]
		arch := mux.Vars(r)["arch"]

		if err := p.driver.Provider.SavePlatformBinary(r.Context(), namespace, registryName, version, os, arch, r.Body); err != nil {
			return err
		}
		defer r.Body.Close()
		return nil
	})
}

func (p *Provider) UploadSHASums() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		if err := p.driver.Provider.SaveSHASUMs(r.Context(), namespace, registryName, version, r.Body); err != nil {
			return err
		}
		defer r.Body.Close()
		return nil
	})
}

func (p *Provider) UploadSHASumsSignature() http.Handler {
	return handler.NewHandler(p.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		registryName := mux.Vars(r)["registryName"]
		version := mux.Vars(r)["version"]

		if err := p.driver.Provider.SaveSHASUMsSig(r.Context(), namespace, registryName, version, r.Body); err != nil {
			return err
		}
		defer r.Body.Close()
		return nil
	})
}
