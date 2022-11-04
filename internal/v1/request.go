package v1

import "github.com/kerraform/kegistry/internal/v1/request"

type AddGPGKeyRequestAttributes struct {
	Namespace  string `json:"namespace" validate:"required"`
	ASCIIArmor string `json:"ascii-armor" validate:"required"`
}

type AddGPGKeyRequest = request.Request[AddGPGKeyRequestAttributes, DataType]
