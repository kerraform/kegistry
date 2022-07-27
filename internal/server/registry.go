package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kerraform/kegistry/internal/handler"
	"github.com/kerraform/kegistry/internal/middleware"
)

const (
	registryPath = "/registry"
)

var (
	v1ModulesPath   = "/v1/modules"
	v1ProvidersPath = "/v1/providers"
)

func (s *Server) registerRegistryHandler() {
	// Service discover
	// See: https://www.terraform.io/internals/remote-service-discovery
	s.mux.Methods(http.MethodGet).Path("/.well-known/terraform.json").Handler(s.ServiceDiscovery())

	registry := s.mux.PathPrefix(registryPath).Subrouter()

	registry.Use(middleware.AccessLog(s.logger))
	registry.Use(middleware.AccessMetric(s.metric))

	// Module Registry Protocol
	// List Available Versions
	// https://www.terraform.io/registry/api-docs#list-available-versions-for-a-specific-module
	registry.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{name}/{provider}/versions", v1ModulesPath)).Handler(s.v1.Module.ListAvailableVersions())

	// Download source code
	// https://www.terraform.io/internals/module-registry-protocol#download-source-code-for-a-specific-module-version
	registry.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{name}/{provider}/{version}/download", v1ModulesPath)).Handler(s.v1.Module.FindSourceCode())

	// Download source code
	// https://www.terraform.io/internals/module-registry-protocol#download-source-code-for-a-specific-module-version
	registry.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{name}/{provider}/{version}/package", v1ModulesPath)).Handler(s.v1.Module.Download())

	// Provider Registry Protocol
	// List Available Versions
	// https://www.terraform.io/internals/provider-registry-protocol#list-available-versions
	registry.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions", v1ProvidersPath)).Handler(s.v1.Provider.ListAvailableVersions())

	// Creates a provider
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#create-a-provider
	registry.Methods(http.MethodPost).Path(v1ProvidersPath).Handler(s.v1.Provider.CreateProvider())

	// Creates a provider version
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	registry.Methods(http.MethodPost).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions", v1ProvidersPath)).Handler(s.v1.Provider.CreateProviderVersion())

	// Creates a provider platform binary
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	registry.Methods(http.MethodPut).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/{os}/{arch}/binary", v1ProvidersPath)).Handler(s.v1.Provider.UploadPlatformBinary())
	registry.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/{os}/{arch}/binary", v1ProvidersPath)).Handler(s.v1.Provider.DownloadPlatformBinary())

	// Creates and get a provider version shasums
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	registry.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/shasums", v1ProvidersPath)).Handler(s.v1.Provider.DownloadSHASums())
	registry.Methods(http.MethodPut).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/shasums", v1ProvidersPath)).Handler(s.v1.Provider.UploadSHASums())

	// Creates and get a provider version shasums signature
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	registry.Methods(http.MethodPut).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/shasums-sig", v1ProvidersPath)).Handler(s.v1.Provider.UploadSHASumsSignature())
	registry.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/shasums-sig", v1ProvidersPath)).Handler(s.v1.Provider.DownloadSHASumsSignature())

	// Creates a provider platform
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-platform
	registry.Methods(http.MethodPost).Path(fmt.Sprintf("%s/{namespace}/{registryName}/versions/{version}/platforms", v1ProvidersPath)).Handler(s.v1.Provider.CreateProviderPlatform())

	// Find a Provider Package
	// https://www.terraform.io/internals/provider-registry-protocol#find-a-provider-package
	registry.Methods(http.MethodGet).Path(fmt.Sprintf("%s/{namespace}/{registryName}/{version}/download/{os}/{arch}", v1ProvidersPath)).Handler(s.v1.Provider.FindPackage())

	// Add GPG Key
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/gpg-keys#add-a-gpg-key
	registry.Methods(http.MethodPost).Path("/v1/gpg-key").Handler(s.v1.AddGPGKey())
}

type GetServiceDiscoveryResponse struct {
	ModulesV1   string `json:"modules.v1"`
	ProvidersV1 string `json:"providers.v1"`
}

func (s *Server) ServiceDiscovery() http.Handler {
	return handler.NewHandler(s.logger, func(w http.ResponseWriter, _ *http.Request) error {
		resp := &GetServiceDiscoveryResponse{
			ModulesV1:   fmt.Sprintf("%s/%s/", registryPath, v1ModulesPath),
			ProvidersV1: fmt.Sprintf("%s/%s/", registryPath, v1ProvidersPath),
		}

		return json.NewEncoder(w).Encode(resp)
	})
}
