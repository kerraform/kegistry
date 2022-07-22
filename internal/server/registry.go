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
	s.mux.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions", v1ProviderPath)).Handler(s.v1.Provider.ListAvailableVersions())

	// Creates a provider
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#create-a-provider
	s.mux.Methods(http.MethodPost).Path(fmt.Sprintf("%s", v1ProviderPath)).Handler(s.v1.Provider.CreateProvider())

	// Creates a provider version
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	s.mux.Methods(http.MethodPost).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions", v1ProviderPath)).Handler(s.v1.Provider.CreateProviderVersion())

	// Creates a provider version shasums
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	s.mux.Methods(http.MethodPost).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/shasums", v1ProviderPath)).Handler(s.v1.Provider.UploadSHASums())

	// Creates a provider version shasums signature
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	s.mux.Methods(http.MethodPost).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/shasums-sig", v1ProviderPath)).Handler(s.v1.Provider.UploadSHASumsSignature())

	// Creates a provider platform
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-platform
	s.mux.Methods(http.MethodPost).Path(fmt.Sprintf("%s/{namespace}/{name}/versions/{version}/platforms", v1ProviderPath)).Handler(s.v1.Provider.CreateProviderPlatform())

	// Find a Provider Package
	// https://www.terraform.io/internals/provider-registry-protocol#find-a-provider-package
	s.mux.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{registryName}/{version}/download/{os}/{arch}", v1ProviderPath)).Handler(s.v1.Provider.FindPackage())

	// Add GPG Key
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/gpg-keys#add-a-gpg-key
	s.mux.Methods(http.MethodPost).Path("/v1/gpg-key").Handler(s.v1.AddGPGKey())
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

		return json.NewEncoder(w).Encode(resp)
	})
}
