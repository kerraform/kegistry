package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"github.com/kerraform/kegistry/internal/v1/module"
	"github.com/kerraform/kegistry/internal/v1/provider"
	"go.uber.org/zap"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

type Handler struct {
	logger *zap.Logger
	driver driver.Driver

	Module   module.Module
	Provider provider.Provider
}

type HandlerConfig struct {
	Driver driver.Driver
	Logger *zap.Logger
}

func New(cfg *HandlerConfig) *Handler {
	module := module.New(&module.Config{
		Driver: cfg.Driver,
		Logger: cfg.Logger.Named("v1.module"),
	})

	provider := provider.New(&provider.Config{
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
	Attributes *AddGPGKeyRequestAttributes `json:"attributes"`
	Type       string                      `json:"type"`
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

		b := bytes.NewBufferString(req.Data.Attributes.ASCIIArmor)
		block, err := armor.Decode(b)
		if err != nil {
			return err
		}

		if block.Type != openpgp.PublicKeyType {
			w.WriteHeader(http.StatusBadRequest)
			return fmt.Errorf("key is not public key")
		}

		reader := packet.NewReader(block.Body)
		pkt, err := reader.Next()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return err
		}

		pgpKey, ok := pkt.(*packet.PublicKey)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return fmt.Errorf("failed to read public key")
		}

		l.Info("received public key",
			zap.String("keyID", pgpKey.KeyIdString()),
		)

		if err := h.driver.SaveGPGKey(r.Context(), req.Data.Attributes.Namespace, pgpKey.KeyIdString(), []byte(req.Data.Attributes.ASCIIArmor)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		defer r.Body.Close()

		l.Info("saved gpg key")
		return nil
	})
}
