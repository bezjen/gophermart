package middleware

import (
	"github.com/bezjen/gophermart/internal/compress"
	"github.com/bezjen/gophermart/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type GzipMiddleware struct {
	logger *logger.Logger
}

func NewGzipMiddleware(logger *logger.Logger) *GzipMiddleware {
	return &GzipMiddleware{logger: logger}
}

func (m *GzipMiddleware) WithGzipRequestDecompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gr, err := compress.NewGzipReader(r.Body)
		if err != nil {
			m.logger.Error("Failed to decompress request body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(gr *compress.GzipReader) {
			err = gr.Close()
			if err != nil {
				m.logger.Error("Failed to close gzip reader", zap.Error(err))
			}
		}(gr)
		r.Body = gr
		h.ServeHTTP(w, r)
	})
}

func (m *GzipMiddleware) WithGzipResponseCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gw := compress.NewGzipWriter(w)
		defer func(gw *compress.GzipWriter) {
			err := gw.Close()
			if err != nil {
				m.logger.Error("Failed to close gzip writer", zap.Error(err))
			}
		}(gw)
		h.ServeHTTP(gw, r)
	})
}
