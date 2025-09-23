package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/bezjen/gophermart/internal/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockHandler struct {
	statusCode int
	response   string
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(m.statusCode)
	w.Write([]byte(m.response))
}

func TestWithGzipRequestDecompression(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		compress      bool
		expectedError bool
	}{
		{
			name:          "normal request",
			content:       "plain text",
			compress:      false,
			expectedError: false,
		},
		{
			name:          "gzipped request",
			content:       "compressed text",
			compress:      true,
			expectedError: false,
		},
	}
	testLogger, _ := logger.NewLogger("debug")
	m := NewGzipMiddleware(testLogger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader = strings.NewReader(tt.content)
			if tt.compress {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				gz.Write([]byte(tt.content))
				gz.Close()
				body = &buf
			}

			req := httptest.NewRequest("POST", "http://localhost:8080", body)
			if tt.compress {
				req.Header.Set("Content-Encoding", "gzip")
			}

			rec := httptest.NewRecorder()
			handler := &mockHandler{statusCode: http.StatusOK, response: "response"}
			wrapped := m.WithGzipRequestDecompression(handler)

			wrapped.ServeHTTP(rec, req)

			if tt.expectedError && rec.Code != http.StatusInternalServerError {
				t.Errorf("expected error status code")
			}
		})
	}
}

func TestWithGzipResponseCompression(t *testing.T) {
	tests := []struct {
		name             string
		acceptEncoding   string
		expectCompressed bool
	}{
		{
			name:             "accepts gzip",
			acceptEncoding:   "gzip",
			expectCompressed: true,
		},
		{
			name:             "no gzip",
			acceptEncoding:   "",
			expectCompressed: false,
		},
	}
	testLogger, _ := logger.NewLogger("debug")
	m := NewGzipMiddleware(testLogger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://localhost:8080", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			rec := httptest.NewRecorder()
			handler := &mockHandler{statusCode: http.StatusOK, response: "response"}
			wrapped := m.WithGzipResponseCompression(handler)

			wrapped.ServeHTTP(rec, req)

			if tt.expectCompressed {
				if rec.Header().Get("Content-Encoding") != "gzip" {
					t.Error("response should be gzipped")
				}
				gr, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatal("failed to create gzip reader")
				}
				defer gr.Close()
				_, err = io.ReadAll(gr)
				if err != nil {
					t.Error("failed to read gzipped content")
				}
			} else {
				if rec.Header().Get("Content-Encoding") == "gzip" {
					t.Error("response should not be gzipped")
				}
				if rec.Body.String() != "response" {
					t.Error("unexpected response body")
				}
			}
		})
	}
}
