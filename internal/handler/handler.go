package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/kerraform/kegistry/internal/logging"
	"go.uber.org/zap"
)

type Error struct {
	Message string `json:"message"`
}

type Handler struct {
	HandleFunc HandlerFunc
}

// HandlerFunc represents the registry handler
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP Implements the http.Handler
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.HandleFunc(w, r)
	if err == nil {
		return
	}

	e := &Error{
		Message: err.Error(),
	}

	l, err := logging.FromCtx(r.Context())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get logger; %v", err)
		return
	}

	if err := json.NewEncoder(w).Encode(e); err != nil {
		l.Error("error to response", zap.Error(err))
		return
	}
	l.Error("error handling request", zap.Error(err))
}

func NewHandler(f HandlerFunc) http.Handler {
	return &Handler{
		HandleFunc: f,
	}
}
