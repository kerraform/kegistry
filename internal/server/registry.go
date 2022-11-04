package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kerraform/kegistry/internal/grammar"
	"github.com/kerraform/kegistry/internal/handler"
	"github.com/kerraform/kegistry/internal/middleware"
	"github.com/kerraform/kegistry/internal/model"
)

const (
	registryPath = "/registry"
)

var (
	v1ModulesPath   = "/v1/modules"
	v1ProvidersPath = "/v1/providers"
)

func (s *Server) registerRegistryHandler() {
	s.mux.Use(middleware.JSON())
	s.mux.Use(middleware.AccessLog(s.logger))
	s.mux.Use(middleware.AccessMetric(s.metric))

	// Service discover
	// See: https://www.terraform.io/internals/remote-service-discovery
	s.mux.Methods(http.MethodGet).Path("/.well-known/terraform.json").Handler(s.ServiceDiscovery())

	registry := s.mux.PathPrefix(registryPath).Subrouter()

	// Add GPG Key
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/gpg-keys#add-a-gpg-key
	registry.Methods(http.MethodPost).Path("/v1/gpg-key").Handler(s.v1.AddGPGKey())

	module := registry.PathPrefix(v1ModulesPath).Subrouter()
	module.Use(middleware.Enable(middleware.ModuleRegistryType, s.enableModule))

	// Module Registry Protocol
	// List Available Versions
	// https://www.terraform.io/registry/api-docs#list-available-versions-for-a-specific-module
	module.Methods(http.MethodGet).Path("/{namespace}/{name}/{provider}/versions").Handler(s.v1.Module.ListAvailableVersions())
	// Create module version
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/modules#create-a-module-version
	module.Methods(http.MethodPost).Path("/{namespace}/{name}/{provider}/versions").Handler(s.v1.Module.CreateModuleVersion())

	// Upload a version
	module.Methods(http.MethodPut).Path(fmt.Sprintf("/{namespace}/{name}/{provider}/versions/{version:%s}", grammar.Version)).Handler(s.v1.Module.UploadModuleVersion())

	// Create module
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/modules#create-a-module-with-no-vcs-connection
	module.Methods(http.MethodPost).Path("/{namespace}").Handler(s.v1.Module.CreateModule())

	// Download source code
	// https://www.terraform.io/internals/module-registry-protocol#download-source-code-for-a-specific-module-version
	module.Methods(http.MethodGet).Path(fmt.Sprintf("/{namespace}/{name}/{provider}/{version:%s}/download", grammar.Version)).Handler(s.v1.Module.FindSourceCode())

	// Download source code
	// https://www.terraform.io/internals/module-registry-protocol#download-source-code-for-a-specific-module-version
	module.Methods(http.MethodGet).Path(fmt.Sprintf("/{namespace}/{name}/{provider}/{version:%s}/{file}", grammar.Version)).Handler(s.v1.Module.Download())

	provider := registry.PathPrefix(v1ProvidersPath).Subrouter()
	provider.Use(middleware.Enable(middleware.ProviderRegistryType, s.enableProvider))

	// Provider Registry Protocol
	// List Available Versions
	// https://www.terraform.io/internals/provider-registry-protocol#list-available-versions
	provider.Methods(http.MethodGet).Path("/{namespace}/{registryName}/versions").Handler(s.v1.Provider.ListAvailableVersions())

	// Creates a provider
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#create-a-provider
	provider.Methods(http.MethodPost).Path("").Handler(s.v1.Provider.CreateProvider())

	// Creates a provider version
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	provider.Methods(http.MethodPost).Path("/{namespace}/{registryName}/versions").Handler(s.v1.Provider.CreateProviderVersion())

	// Creates a provider platform binary
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	provider.Methods(http.MethodPut).Path(fmt.Sprintf("/{namespace}/{registryName}/versions/{version:%s}/{os}/{arch}/binary", grammar.Version)).Handler(s.v1.Provider.UploadPlatformBinary())
	provider.Methods(http.MethodGet).Path(fmt.Sprintf("/{namespace}/{registryName}/versions/{version:%s}/{os}/{arch}/binary", grammar.Version)).Handler(s.v1.Provider.DownloadPlatformBinary())

	// Creates and get a provider version shasums
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	provider.Methods(http.MethodGet).Path(fmt.Sprintf("/{namespace}/{registryName}/versions/{version:%s}/shasums", grammar.Version)).Handler(s.v1.Provider.DownloadSHASums())
	provider.Methods(http.MethodPut).Path(fmt.Sprintf("/{namespace}/{registryName}/versions/{version:%s}/shasums", grammar.Version)).Handler(s.v1.Provider.UploadSHASums())

	// Creates and get a provider version shasums signature
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
	provider.Methods(http.MethodPut).Path(fmt.Sprintf("/{namespace}/{registryName}/versions/{version:%s}/shasums-sig", grammar.Version)).Handler(s.v1.Provider.UploadSHASumsSignature())
	provider.Methods(http.MethodGet).Path(fmt.Sprintf("/{namespace}/{registryName}/versions/{version:%s}/shasums-sig", grammar.Version)).Handler(s.v1.Provider.DownloadSHASumsSignature())

	// Creates a provider platform
	// Inspired by Terraform Cloud API:
	// https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-platform
	provider.Methods(http.MethodPost).Path(fmt.Sprintf("/{namespace}/{registryName}/versions/{version:%s}/platforms", grammar.Version)).Handler(s.v1.Provider.CreateProviderPlatform())

	// Find a Provider Package
	// https://www.terraform.io/internals/provider-registry-protocol#find-a-provider-package
	provider.Methods(http.MethodGet).Path(fmt.Sprintf("/{namespace}/{registryName}/{version:%s}/download/{os}/{arch}", grammar.Version)).Handler(s.v1.Provider.FindPackage())
}

func (s *Server) ServiceDiscovery() http.Handler {
	return handler.NewHandler(func(w http.ResponseWriter, _ *http.Request) error {
		resp := &model.Service{
			ModulesV1:   registryPath + v1ModulesPath + "/",
			ProvidersV1: registryPath + v1ProvidersPath + "/",
		}

		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(resp)
	})
}
