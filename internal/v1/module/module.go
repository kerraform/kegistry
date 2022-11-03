package module

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"go.uber.org/zap"
)

type DataType string

const (
	DataTypeRegistryModule        DataType = "registry-modules"
	DataTypeRegistryModuleVersion DataType = "registry-module-versions"
)

type Module struct {
	driver *driver.Driver
	logger *zap.Logger
}

type Config struct {
	Driver *driver.Driver
	Logger *zap.Logger
}

func New(cfg *Config) *Module {
	return &Module{
		driver: cfg.Driver,
		logger: cfg.Logger,
	}
}

type CreateModuleRequest struct {
	Data *CreateModuleRequestData `json:"data"`
}

type CreateModuleRequestData struct {
	Attributes *CreateModuleDataAttributes `json:"attributes"`
	Type       DataType                    `json:"type"`
}

type CreateModuleDataAttributes struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

func (m *Module) CreateModule() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]

		var req CreateModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		defer r.Body.Close()

		if err := m.driver.Module.CreateModule(r.Context(), namespace, req.Data.Attributes.Provider, req.Data.Attributes.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

type CreateModuleVersionRequest struct {
	Data *CreateModuleVersionRequestData `json:"data"`
}

type CreateModuleVersionRequestData struct {
	Attributes *CreateModuleVersionDataAttributes `json:"attributes"`
	Type       DataType                           `json:"type"`
}

type CreateModuleVersionDataAttributes struct {
	Version string `json:"version"`
}

type CreateModuleVersionResponse struct {
	Data *CreateModuleVersionData `json:"data"`
}

type CreateModuleVersionData struct {
	Links *CreateModuleVersionDataLinks `json:"links"`
}

type CreateModuleVersionDataLinks struct {
	Upload string `json:"upload"`
}

func (m *Module) CreateModuleVersion() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		provider := mux.Vars(r)["provider"]
		name := mux.Vars(r)["name"]

		var req CreateModuleVersionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		result, err := m.driver.Module.CreateVersion(r.Context(), namespace, provider, name, req.Data.Attributes.Version)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		resp := &CreateModuleVersionResponse{
			Data: &CreateModuleVersionData{
				Links: &CreateModuleVersionDataLinks{
					Upload: result.Upload,
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(resp)
	})
}

func (m *Module) Download() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		name := mux.Vars(r)["name"]
		provider := mux.Vars(r)["provider"]
		version := mux.Vars(r)["version"]

		f, err := m.driver.Module.GetModule(r.Context(), namespace, name, provider, version)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				return driver.ErrModuleNotExist
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := io.Copy(w, f); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		defer f.Close()

		w.WriteHeader(http.StatusOK)
		return nil
	})
}

// https://www.terraform.io/registry/api-docs#download-source-code-for-a-specific-module-version
func (m *Module) FindSourceCode() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		name := mux.Vars(r)["name"]
		provider := mux.Vars(r)["provider"]
		version := mux.Vars(r)["version"]

		url, err := m.driver.Module.GetDownloadURL(r.Context(), namespace, provider, name, version)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				return driver.ErrModuleNotExist
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		w.Header().Set("X-Terraform-Get", url)
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

type ListAvailableVersionsResponse struct {
	Modules []ListAvailableVersionsModel `json:"modules"`
}

type ListAvailableVersionsModel struct {
	Versions []ListAvailableVersionsModelVersion `json:"versions"`
}

type ListAvailableVersionsModelVersion struct {
	Version string `json:"version"`
}

// https://www.terraform.io/internals/module-registry-protocol#list-available-versions-for-a-specific-module
func (m *Module) ListAvailableVersions() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		name := mux.Vars(r)["name"]
		provider := mux.Vars(r)["provider"]

		versions, err := m.driver.Module.ListAvailableVersions(r.Context(), namespace, provider, name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		vs := make([]ListAvailableVersionsModelVersion, len(versions))
		for i, version := range versions {
			vs[i] = ListAvailableVersionsModelVersion{
				Version: version,
			}
		}

		resp := &ListAvailableVersionsResponse{
			Modules: []ListAvailableVersionsModel{
				{
					Versions: vs,
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(resp)
	})
}

func (m *Module) UploadModuleVersion() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		provider := mux.Vars(r)["provider"]
		name := mux.Vars(r)["name"]
		version := mux.Vars(r)["version"]

		if err := m.driver.Module.SavePackage(r.Context(), namespace, provider, name, version, r.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		defer r.Body.Close()

		w.WriteHeader(http.StatusOK)
		return nil
	})
}
