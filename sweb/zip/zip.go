package zip

import (
	"compress/flate"
	"compress/gzip"
	"net/http"
	"strings"
)

const (
	ENCODING_GZIP    = "gzip"
	ENCODING_DEFLATE = "deflate"

	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerContentLength   = "Content-Length"
	headerContentType     = "Content-Type"
	headerVary            = "Vary"
	headerSecWebSocketKey = "Sec-WebSocket-Key"
)

//first gzip
func GetZipAcceptEncoding(r *http.Request) string {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, ENCODING_GZIP) {
		return ENCODING_GZIP
	} else if strings.Contains(acceptEncoding, ENCODING_DEFLATE) {
		return ENCODING_DEFLATE
	} else {
		return ""
	}
}

func CanZip(w http.ResponseWriter, r *http.Request) bool {
	// Skip compression if the client doesn't accept gzip encoding.
	if len(r.Header.Get(headerSecWebSocketKey)) > 0 {
		return false
	}

	// Skip compression if already comprssed
	writerencoding := w.Header().Get(headerContentEncoding)
	if writerencoding == ENCODING_GZIP || writerencoding == ENCODING_DEFLATE {
		return false
	}
	return true
}

func GzipWrite(w http.ResponseWriter, level int, b []byte) (int, error) {
	gz, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return 0, err
	}
	return gzipWrite(w, gz, b)
}

func gzipWrite(w http.ResponseWriter, gz *gzip.Writer, b []byte) (int, error) {
	if len(w.Header().Get(headerContentType)) == 0 {
		w.Header().Set(headerContentType, http.DetectContentType(b))
	}
	w.Header().Set(headerContentEncoding, ENCODING_GZIP)
	w.Header().Set(headerVary, headerAcceptEncoding)
	n, err := gz.Write(b)
	if err == nil {
		w.Header().Del(headerContentLength)
	}
	gz.Close()
	return n, err
}

func NewGzipWriter(w http.ResponseWriter, level int) (http.ResponseWriter, error) {
	gz, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return nil, err
	}

	return &gzipResponseWriter{
		gzipw:          gz,
		ResponseWriter: w,
	}, nil
}

func deflateWrite(w http.ResponseWriter, dw *flate.Writer, b []byte) (int, error) {
	if len(w.Header().Get(headerContentType)) == 0 {
		w.Header().Set(headerContentType, http.DetectContentType(b))
	}
	w.Header().Set(headerContentEncoding, ENCODING_DEFLATE)
	w.Header().Set(headerVary, headerAcceptEncoding)
	n, err := dw.Write(b)
	if err == nil {
		w.Header().Del(headerContentLength)
	}
	dw.Close()
	return n, err
}

func DeflateWrite(w http.ResponseWriter, level int, b []byte) (int, error) {
	deflate, err := flate.NewWriter(w, level)
	if err != nil {
		return 0, err
	}
	return deflateWrite(w, deflate, b)
}

func NewDefalteWriter(w http.ResponseWriter, level int) (http.ResponseWriter, error) {
	deflate, err := flate.NewWriter(w, level)
	if err != nil {
		return nil, err
	}

	return &deflateResponseWriter{
		deflatew:       deflate,
		ResponseWriter: w,
	}, nil
}

type gzipResponseWriter struct {
	gzipw *gzip.Writer
	http.ResponseWriter
}

func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	return gzipWrite(grw.ResponseWriter, grw.gzipw, b)
}

func (grw gzipResponseWriter) Close() {
	grw.gzipw.Close()
}

type deflateResponseWriter struct {
	deflatew *flate.Writer
	http.ResponseWriter
}

func (drw deflateResponseWriter) Write(b []byte) (int, error) {
	return deflateWrite(drw.ResponseWriter, drw.deflatew, b)
}

func (drw deflateResponseWriter) Close() {
	drw.deflatew.Close()
}
