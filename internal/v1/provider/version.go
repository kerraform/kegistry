package provider

import model "github.com/kerraform/kegistry/internal/model/provider"

type AvailableVersion struct {
	Version   string                           `json:"version"`
	Protocols []string                         `json:"protocols"`
	Platforms []model.AvailableVersionPlatform `json:"platforms"`
}
