package static

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	app_context "github.com/morphy76/g-fe-server/internal/http/context"
)

func HandleStatic(staticRouter *mux.Router, context context.Context) {

	ctxRoot := context.Value(app_context.CTX_CONTEXT_ROOT_KEY).(app_context.ServeOptions).ContextRoot
	staticPath := context.Value(app_context.CTX_CONTEXT_ROOT_KEY).(app_context.ServeOptions).StaticPath

	defaultFile := filepath.Join(staticPath, "index.html")

	fileServer := func(w http.ResponseWriter, r *http.Request) {

		path := filepath.Join(staticPath, strings.TrimPrefix(r.URL.Path, ctxRoot+"/ui"))

		fi, err := os.Stat(path)

		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if fi.IsDir() {
			http.ServeFile(w, r, defaultFile)
		} else {
			http.ServeFile(w, r, path)
		}
	}

	staticRouter.Methods(http.MethodGet).HandlerFunc(fileServer).Name(ctxRoot + "/ui")
}
