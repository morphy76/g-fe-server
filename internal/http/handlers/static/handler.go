package static

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

func HandleStatic(staticRouter *mux.Router, ctxRoot string, staticPath string) {

	defaultFile := filepath.Join(staticPath, "index.html")

	fileServer := func(w http.ResponseWriter, r *http.Request) {

		requestedFile := filepath.Join(staticPath, strings.TrimPrefix(r.URL.Path, ctxRoot+"/ui"))

		requestedFileStats, err := os.Stat(requestedFile)
		if os.IsNotExist(err) {
			requestedFile = defaultFile
			requestedFileStats, _ = os.Stat(requestedFile)
		}

		if !strings.HasSuffix(requestedFile, ".js") {
			w.Header().Set("Cache-Control", "no-cache")
		}

		if requestedFileStats.IsDir() {
			http.ServeFile(w, r, defaultFile)
		} else {
			http.ServeFile(w, r, requestedFile)
		}
	}

	staticRouter.Methods(http.MethodGet).HandlerFunc(fileServer).Name("GET " + ctxRoot + "/ui")
}
