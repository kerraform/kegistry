package provider

import "github.com/kerraform/kegistry/internal/v1/request"

// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#request-body
type CreateProviderRequest = request.Request[CreateProviderRequestDataAttributes, DataType]
type CreateProviderPlatformRequest = request.Request[CreateProviderPlatformRequestDataAttributes, DataType]

// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#request-body
type CreateProviderVersionRequest = request.Request[CreateProviderVersionRequestDataAttributes, DataType]
