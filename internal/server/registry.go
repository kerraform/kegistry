package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/kerraform/kegistry/internal/handler"
)

const (
	v1ModulePath   = "/v1/modules"
	v1ProviderPath = "/v1/providers"
)

func (s *Server) registerRegistryHandler() {
	// Service discover
	// See: https://www.terraform.io/internals/remote-service-discovery
	s.mux.Methods(http.MethodGet).Path("/.well-known/terraform.json").Handler(s.ServiceDiscovery())

	// Module Registry Protocol
	// List Available Versions
	// https://www.terraform.io/internals/module-registry-protocol#list-available-versions-for-a-specific-module
	s.mux.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{name}/{provider}/versions", v1ModulePath)).Handler(s.v1.Provider.ListAvailableVersions())

	// Download source code
	// https://www.terraform.io/internals/module-registry-protocol#download-source-code-for-a-specific-module-version
	s.mux.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{name}/{system}/{version}/download", v1ModulePath)).Handler(s.v1.Provider.FindPackage())

	// Provider Registry Protocol
	// List Available Versions
	// https://www.terraform.io/internals/provider-registry-protocol#list-available-versions
	s.mux.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{type}/versions", v1ProviderPath)).Handler(s.v1.Provider.ListAvailableVersions())

	// Find a Provider Package
	// https://www.terraform.io/internals/provider-registry-protocol#find-a-provider-package
	s.mux.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{type}/{version}/download/{os}/{arch}", v1ProviderPath)).Handler(s.v1.Provider.FindPackage())
}

type GetServiceDiscoveryResponse struct {
	ModuleV1   string `json:"module.v1"`
	ProviderV1 string `json:"provider.v1"`
}

func (s *Server) ServiceDiscovery() http.Handler {
	return handler.NewHandler(s.logger, func(w http.ResponseWriter, _ *http.Request) error {
		resp := &GetServiceDiscoveryResponse{
			ModuleV1:   path.Join(s.baseURL.String(), v1ModulePath),
			ProviderV1: path.Join(s.baseURL.String(), v1ProviderPath),
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		return json.NewEncoder(w).Encode(resp)
	})
}
