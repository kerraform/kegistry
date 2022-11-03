package provider

import (
	model "github.com/kerraform/kegistry/internal/model/provider"
)

type Response[T any] struct {
	Data *T `json:"data"`
}

type CreateProviderPlatformResponse = Response[CreateProviderPlatformResponseData]
type CreateProviderVersionResponse = Response[CreateProviderVersionResponseData]

type ListAvailableVersionsResponse struct {
	Versions []model.AvailableVersion `json:"versions"`
}
