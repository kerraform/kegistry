package handler

import (
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	logger     *zap.Logger
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

	h.logger.Error("error handling request", zap.Error(err))
}

func NewHandler(logger *zap.Logger, f HandlerFunc) http.Handler {
	return &Handler{
		logger:     logger,
		HandleFunc: f,
	}
}
