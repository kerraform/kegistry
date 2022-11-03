package provider

type Request[T any, K comparable] struct {
	Data *Data[T, K] `json:"data"`
}

type Data[T any, K comparable] struct {
	Attributes *T `json:"attributes"`
	Type       K  `json:"type"`
}

// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#request-body
type CreateProviderRequest = Request[CreateProviderRequestDataAttributes, DataType]
type CreateProviderPlatformRequest = Request[CreateProviderPlatformRequestDataAttributes, DataType]

// https://www.terraform.io/cloud-docs/api-docs/private-registry/providers#request-body
type CreateProviderVersionRequest = Request[CreateProviderVersionRequestDataAttributes, DataType]
