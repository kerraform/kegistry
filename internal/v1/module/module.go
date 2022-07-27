package module

import (
	"encoding/json"
	"net/http"

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
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("X-Terraform-Get", "")
		return nil
	})
}

func (m *module) Download() http.Handler {
	return handler.NewHandler(m.logger, func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("X-Terraform-Get", "")
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
