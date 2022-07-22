package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"go.uber.org/zap"
)

type Handler struct {
	logger *zap.Logger
	driver driver.Driver

	Module   Module
	Provider Provider
}

type HandlerConfig struct {
	Driver driver.Driver
	Logger *zap.Logger
}

func New(cfg *HandlerConfig) *Handler {
	module := newModule(&moduleConfig{
		Driver: cfg.Driver,
		Logger: cfg.Logger.Named("v1.module"),
	})

	provider := newProvider(&providerConfig{
		Driver: cfg.Driver,
		Logger: cfg.Logger.Named("v1.provider"),
	})

	return &Handler{
		driver:   cfg.Driver,
		logger:   cfg.Logger.Named("v1"),
		Module:   module,
		Provider: provider,
	}
}

type AddGPGKeyRequest struct {
	Data *AddGPGKeyRequestData `json:"data"`
}

type AddGPGKeyRequestData struct {
	Type       string                      `json:"type"`
	Attributes *AddGPGKeyRequestAttributes `json:"attributes"`
}

type AddGPGKeyRequestAttributes struct {
	Namespace  string `json:"namespace"`
	ASCIIArmor string `json:"ascii-armor"`
}

func (req *AddGPGKeyRequest) Valid() bool {
	return req.Data.Attributes.Namespace != "" && req.Data.Attributes.ASCIIArmor != ""
}

func (h *Handler) AddGPGKey() http.Handler {
	return handler.NewHandler(h.logger, func(w http.ResponseWriter, r *http.Request) error {
		var req AddGPGKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil
		}
		defer r.Body.Close()

		l := h.logger.With(
			zap.String("namespace", req.Data.Attributes.Namespace),
		)

		if valid := req.Valid(); !valid {
			w.WriteHeader(http.StatusBadRequest)
			return fmt.Errorf("invalid request")
		}

		if err := h.driver.SaveGPGKey(req.Data.Attributes.Namespace, req.Data.Attributes.ASCIIArmor); err != nil {
			return err
		}
		l.Info("saved gpg key")
		return nil
	})
}
