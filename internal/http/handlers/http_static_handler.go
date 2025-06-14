package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

const contentTypeHTTPHeader = "Content-Type"

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
			ext := strings.ToLower(filepath.Ext(requestedFile))
			switch ext {
			case ".css":
				w.Header().Set(contentTypeHTTPHeader, "text/css")
			case ".js":
				w.Header().Set(contentTypeHTTPHeader, "application/javascript")
			case ".html":
				w.Header().Set(contentTypeHTTPHeader, "text/html")
			case ".png":
				w.Header().Set(contentTypeHTTPHeader, "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set(contentTypeHTTPHeader, "image/jpeg")
			case ".gif":
				w.Header().Set(contentTypeHTTPHeader, "image/gif")
			case ".svg":
				w.Header().Set(contentTypeHTTPHeader, "image/svg+xml")
			default:
				w.Header().Set(contentTypeHTTPHeader, "application/octet-stream")
			}
			http.ServeFile(w, r, requestedFile)
		}
	}

	staticRouter.Methods(http.MethodGet).HandlerFunc(fileServer).Name("GET " + ctxRoot + "/ui")

	return nil
}
