package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzipWriter struct {
	rw http.ResponseWriter
	gw *gzip.Writer
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	return &GzipWriter{
		rw: w,
		gw: gzip.NewWriter(w),
	}
}

func (w GzipWriter) Header() http.Header {
	return w.rw.Header()
}

func (w GzipWriter) Write(p []byte) (int, error) {
	return w.gw.Write(p)
}

func (w GzipWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusNoContent || statusCode == http.StatusConflict {
		w.rw.Header().Set("Content-Encoding", "gzip")
	}
	w.rw.WriteHeader(statusCode)
}

func (w GzipWriter) Close() error {
	return w.gw.Close()
}

type GzipReader struct {
	rc io.ReadCloser
	gr *gzip.Reader
}

func NewGzipReader(r io.ReadCloser) (*GzipReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &GzipReader{
		rc: r,
		gr: gr,
	}, nil
}

func (r GzipReader) Read(p []byte) (n int, err error) {
	return r.gr.Read(p)
}

func (r GzipReader) Close() error {
	if err := r.rc.Close(); err != nil {
		return err
	}
	return r.gr.Close()
}
