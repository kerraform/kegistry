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

type Module interface {
	Download() http.Handler
	FindSourceCode() http.Handler
	ListAvailableVersions() http.Handler
}

type module struct {
	driver *driver.Driver
	logger *zap.Logger
}

var _ Module = (*module)(nil)

type Config struct {
	Driver *driver.Driver
	Logger *zap.Logger
}

func New(cfg *Config) Module {
	return &module{
		driver: cfg.Driver,
		logger: cfg.Logger,
	}
}

// https://www.terraform.io/registry/api-docs#download-source-code-for-a-specific-module-version
func (m *module) FindSourceCode() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		name := mux.Vars(r)["name"]
		provider := mux.Vars(r)["provider"]
		version := mux.Vars(r)["version"]

		l := m.logger.With(
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.String("provider", provider),
			zap.String("version", version),
		)

		url, err := m.driver.Module.GetDownloadURL(r.Context(), namespace, provider, name, version)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				return driver.ErrModuleNotExist
			}

			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		l.Debug("found source code of module",
			zap.String("url", url),
		)
		w.Header().Set("X-Terraform-Get", url)
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func (m *module) Download() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		name := mux.Vars(r)["name"]
		provider := mux.Vars(r)["provider"]
		version := mux.Vars(r)["version"]

		l := m.logger.With(
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.String("provider", provider),
			zap.String("version", version),
		)

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

		l.Debug("distributed module")
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
func (m *module) ListAvailableVersions() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, r *http.Request) error {
		namespace := mux.Vars(r)["namespace"]
		name := mux.Vars(r)["name"]
		provider := mux.Vars(r)["provider"]

		l := m.logger.With(
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.String("provider", provider),
		)

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

		l.Info("list available module versions")
		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(resp)
	})
}
