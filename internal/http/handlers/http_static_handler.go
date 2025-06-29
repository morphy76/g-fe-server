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

		requestedFile := filepath.Join(staticPath, strings.TrimPrefix(r.URL.Path, ctxRoot+"/ui"))

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
