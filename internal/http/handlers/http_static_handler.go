package handlers

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

const contentTypeHTTPHeader = "Content-Type"

// HandleStatic registers a static file server for the given static path.
func HandleStatic(staticRouter *mux.Router, ctxRoot string, staticPath string) error {

	defaultFile := filepath.Join(staticPath, "index.html")

	fileServer := func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if recover := recover(); recover != nil {
				http.NotFound(w, r)
			}
		}()

		requestedPath := strings.TrimPrefix(r.URL.Path, ctxRoot+"/ui")
		requestedPath = filepath.Clean(requestedPath)

		if strings.Contains(requestedPath, "..") || strings.HasPrefix(requestedPath, "/") {
			http.NotFound(w, r)
			return
		}

		requestedFile := filepath.Join(staticPath, requestedPath)

		absStaticPath, err := filepath.Abs(staticPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		absRequestedFile, err := filepath.Abs(requestedFile)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if !strings.HasPrefix(absRequestedFile, absStaticPath) {
			http.NotFound(w, r)
			return
		}

		requestedFileStats, err := os.Stat(requestedFile)
		if os.IsNotExist(err) {
			requestedFile = defaultFile
			requestedFileStats, _ = os.Stat(requestedFile)
		}

		if requestedFileStats.IsDir() {
			w.Header().Set(contentTypeHTTPHeader, "text/html")
			http.ServeFile(w, r, defaultFile)
		} else {
			ext := filepath.Ext(requestedFile)
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			w.Header().Set(contentTypeHTTPHeader, contentType)
			http.ServeFile(w, r, requestedFile)
		}
	}

	staticRouter.Methods(http.MethodGet).HandlerFunc(fileServer).Name("GET " + ctxRoot + "/ui")

	return nil
}
