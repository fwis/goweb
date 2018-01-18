package render

import (
	"compress/flate"
	"compress/gzip"
	"net/http"
)

type HTMLRenderOption struct {
	Zip          bool
	ZipThreshold int
	GzipLevel    int
	DeflateLevel int
	Before       http.HandlerFunc
}

func NewDefaultHTMLRenderOption() *HTMLRenderOption {
	return &HTMLRenderOption{
		Zip:          true,
		ZipThreshold: 512,
		GzipLevel:    gzip.DefaultCompression,
		DeflateLevel: flate.DefaultCompression,
		Before:       defaultBefore,
	}
}

func defaultBefore(w http.ResponseWriter, r *http.Request) {
	//DENY ： 不允许被任何页面嵌入；
	//SAMEORIGIN ： 不允许被本域以外的页面嵌入；
	//ALLOW-FROM uri： 不允许被指定的域名以外的页面嵌入（Chrome现阶段不支持）
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type-Options", "nosniff")
	//m.w.Header().Set("Content-Security-Policy", "default-src 'self'")
}
