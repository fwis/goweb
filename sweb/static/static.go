package static

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func NewStatic() *StaticServer {
	s := &StaticServer{}
	s.url2dir = make(map[string]string)
	s.url2file = make(map[string]string)
	return s
}

type StaticServer struct {
	url2file         map[string]string
	url2dir          map[string]string
	cache_extensions []string
	maxage           int64
}

func (s *StaticServer) Url2File(urlpath string, filepath string) error {
	if strings.HasSuffix(filepath, "/") {
		return fmt.Errorf("Invalid static file URL")
	}

	file_stat, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	if file_stat.IsDir() {
		return fmt.Errorf("Static filepath is dir")
	}

	s.url2file[urlpath] = filepath
	return nil
}

//urlpath must end with "/"
func (s *StaticServer) Url2Dir(urlpath string, dirpath string) error {
	if !strings.HasSuffix(urlpath, "/") {
		return fmt.Errorf("Invalid static dir URL")
	}

	dir_stat, err := os.Stat(dirpath)
	if err != nil {
		return err
	}

	if !dir_stat.IsDir() {
		return fmt.Errorf("Static filepath is NOT dir")
	}

	s.url2dir[urlpath] = dirpath
	return nil
}

//mime.AddExtensionType(".mustache", "text/html; charset=utf-8")
func (s *StaticServer) AddExtensionType(extension string, mimetype string) {
	mime.AddExtensionType(extension, mimetype)
}

func (s *StaticServer) CacheExtensions(extensions []string) {
	s.cache_extensions = extensions
}

//if return false, means find static file
func (s *StaticServer) FilterHTTP(w http.ResponseWriter, r *http.Request) bool {
	fpath := s.matchfile(r)
	if fpath == "" {
		return true
	} else {
		finfo, err := os.Stat(fpath)
		if err != nil {
			http.NotFound(w, r)
			return false
		}

		if finfo.IsDir() {
			http.Error(w, "403 Forbidden", http.StatusForbidden)
			return false
		}

		if s.maxage > 0 {
			w.Header().Set("Cache-Control", "max-age="+strconv.FormatInt(s.maxage, 10))
		}

		if s.canCacheZip(fpath) {
			//fmt.Printf("s.canCacheZip return true,fpath=%s\n", fpath)
			s.cacheZip(fpath, finfo, w, r)
		} else {
			http.ServeFile(w, r, fpath)
		}
	}

	return false
}

func (s *StaticServer) matchfile(r *http.Request) string {
	for url, fpath := range s.url2file {
		if r.URL.Path == url {
			return fpath
		}
	}

	for url_prefix, dir := range s.url2dir {
		if strings.HasPrefix(r.URL.Path, url_prefix) {
			fpath := filepath.Join(dir, r.URL.Path[len(url_prefix):])
			return fpath
		}
	}
	return ""
}

func (s *StaticServer) canCacheZip(fpath string) bool {
	if len(s.cache_extensions) <= 0 {
		return false
	}

	for _, extenstion := range s.cache_extensions {
		if strings.HasSuffix(fpath, extenstion) {
			return true
		}
	}
	return false
}

func WriteContentHead(w http.ResponseWriter, contentEncoding string, contentlength int64) {
	if contentEncoding == "gzip" {
		w.Header().Set("Content-Encoding", "gzip")
	} else if contentEncoding == "deflate" {
		w.Header().Set("Content-Encoding", "deflate")
	} else {
		w.Header().Set("Content-Length", strconv.FormatInt(contentlength, 10))
	}
}

//进入静态文件zip+cache逻辑
//1. 采用了最高压缩率
//2. 将zip后的内容cache, 极大降低 io 和 cpu。 用内存换时间, 要评估好内存大小
//   mem := 每个静态文件(原始大小+最高压缩gzip后大小+deflate压缩后大小) 之和
//3. 支持Last-Modified
//4. 文件改动后, 能更新Last-Modified和cache
func (s *StaticServer) cacheZip(fpath string, finfo os.FileInfo, w http.ResponseWriter, r *http.Request) {
	contentEncoding := GetAcceptEncodingZip(r)

	//如果w.contentEncoding是空, 不压缩
	memzipfile, err := OpenMemZipFile(fpath, contentEncoding)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	//初始化response head content-encoding, content-length
	WriteContentHead(w, contentEncoding, finfo.Size())

	//gzip一个未知的mimetype的内容后,如果不明确设置content-type
	//go会根据内容的头几个字节，自动判断为application/x-gzip
	//这导致浏览器认为是一个zip文件下载。
	//两种方式解决这个问题:
	//1. 调用 mime.AddExtensionType(ext, typ string), 明确告诉go你自己的mimetype
	//   例如 mime.AddExtensionType(".mustache", "text/html; charset=utf-8")
	//2. 修改 /etc/mimetype
	//3. 在这里HardCode
	// if strings.HasSuffix(fpath, ".mustache") {
	// 	w.Header().Set("Content-Type", "text/html; charset=utf-8") //FIXME: hardcode
	// }

	http.ServeContent(w, r, fpath, finfo.ModTime(), memzipfile)
}
