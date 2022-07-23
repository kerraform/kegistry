package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"go.uber.org/zap"
	"golang.org/x/crypto/openpgp"
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

		key, ok := pkt.(*packet.PublicKey)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return fmt.Errorf("failed to read public key")
		}

		l.Info("received public key",
			zap.String("keyID", key.KeyIdString()),
		)

		if err := h.driver.SaveGPGKey(req.Data.Attributes.Namespace, key); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		l.Info("saved gpg key")
		return nil
	})
}
