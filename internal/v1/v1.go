package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/handler"
	"github.com/kerraform/kegistry/internal/logging"
	"github.com/kerraform/kegistry/internal/v1/module"
	"github.com/kerraform/kegistry/internal/v1/provider"
	"github.com/kerraform/kegistry/internal/validator"
	"go.uber.org/zap"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

type DataType string

const (
	DataTypeAddGPGKey DataType = "gpg-keys"
)

type Handler struct {
	logger *zap.Logger
	driver *driver.Driver

	Module   *module.Module
	Provider *provider.Provider
}

type HandlerConfig struct {
	Driver *driver.Driver
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

func (h *Handler) AddGPGKey() http.Handler {
	return handler.NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		var req AddGPGKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil
		}
		defer r.Body.Close()

		l, err := logging.FromCtx(r.Context())
		if err != nil {
			l.Error("failed to get logger", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return errors.New("internal error")
		}

		if req.Data.Type != DataTypeAddGPGKey {
			w.WriteHeader(http.StatusBadRequest)
			return fmt.Errorf("data type is not %s", DataTypeAddGPGKey)
		}

		if err := validator.Validate.Struct(req); err != nil {
			l.Error("failed to validate struct", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return errors.New("invalid request body")
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

		if err := h.driver.Provider.SaveGPGKey(r.Context(), req.Data.Attributes.Namespace, pgpKey.KeyIdString(), []byte(req.Data.Attributes.ASCIIArmor)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		defer r.Body.Close()
		return nil
	})
}
