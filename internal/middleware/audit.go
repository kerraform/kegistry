package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	"go.uber.org/zap"
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

func Audit(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res := &http.Response{}
			rww := newRwWrapper(w, res)

			defer func() {
				logger.Named("accessLog").Info("access to server",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("userAgent", r.UserAgent()),
					zap.String("contentLength", strconv.FormatInt(r.ContentLength, 10)),
					zap.String("query", r.URL.Query().Encode()),
					zap.Int("statusCode", res.StatusCode),
				)
			}()
			next.ServeHTTP(rww, r)
		})
	}
}
