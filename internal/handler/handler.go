package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type Error struct {
	Message string `json:"message"`
}

type Handler struct {
	logger     *zap.Logger
	HandleFunc HandlerFunc
}

// HandlerFunc represents the registry handler
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP Implements the http.Handler
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := h.HandleFunc(w, r)
	if err == nil {
		return
	}

	e := &Error{
		Message: err.Error(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(e); err != nil {
		h.logger.Error("error to response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
	h.logger.Error("error handling request", zap.Error(err))
}

func NewHandler(logger *zap.Logger, f HandlerFunc) http.Handler {
	return &Handler{
		logger:     logger,
		HandleFunc: f,
	}
}
