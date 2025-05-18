package etag

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/fs"
	"log"
	"net/http"
)

func HandlerFunc(fileSystem fs.FS, stripPrefix string) http.HandlerFunc {
	etags := generateETags(fileSystem, stripPrefix)
	handler := http.FileServer(http.FS(fileSystem))
	if stripPrefix != "" {
		handler = http.StripPrefix(stripPrefix, handler)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if sum, ok := etags[r.URL.Path]; ok {
			w.Header().Set("ETag", sum)
			if r.Header.Get("If-None-Match") == sum {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
		handler.ServeHTTP(w, r)
	}
}

func Handler(fileSystem fs.FS, stripPrefix string) http.Handler {
	return http.HandlerFunc(HandlerFunc(fileSystem, stripPrefix))
}

func generateETags(fileSystem fs.FS, stripPrefix string) map[string]string {
	etags := make(map[string]string)
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		in, err := fileSystem.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		hasher := sha1.New()
		if _, err := io.Copy(hasher, in); err != nil {
			return err
		}

		sum := hex.EncodeToString(hasher.Sum(nil))
		etags[stripPrefix+"/"+path] = sum
		return nil
	})
	if err != nil {
		log.Println("[error] ETags:", err.Error())
	}
	return etags
}
