package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type rwWrapper struct {
	rw     http.ResponseWriter
	mirror *http.Response
	closed bool
}

// newRwWrapper wraps the HTTP responseWriter for audit logging
func newRwWrapper(rw http.ResponseWriter, mirror *http.Response) *rwWrapper {
	return &rwWrapper{
		rw:     rw,
		mirror: mirror,
	}
}

func (r *rwWrapper) Header() http.Header {
	return r.rw.Header()
}

func (r *rwWrapper) Write(i []byte) (int, error) {
	r.mirror.Body = ioutil.NopCloser(bytes.NewReader(i))
	return r.rw.Write(i)
}

func (r *rwWrapper) WriteHeader(statusCode int) {
	if r.closed {
		return
	}
	r.closed = true
	r.rw.WriteHeader(statusCode)
	r.mirror.StatusCode = statusCode
}

func AccessLog(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res := &http.Response{}
			rww := newRwWrapper(w, res)

			defer func() {
				logger.Named("accessLog").Info("access to server",
					flattenVars(r, res)...,
				)
			}()
			next.ServeHTTP(rww, r)
		})
	}
}

func flattenVars(r *http.Request, res *http.Response) []zapcore.Field {
	fs := []zapcore.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("userAgent", r.UserAgent()),
		zap.String("contentLength", strconv.FormatInt(r.ContentLength, 10)),
		zap.String("query", r.URL.Query().Encode()),
		zap.Int("statusCode", res.StatusCode),
	}
	for k, v := range mux.Vars(r) {
		fs = append(fs, zap.String(k, v))
	}

	return fs
}
