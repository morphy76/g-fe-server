package static

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

func HandleStatic(staticRouter *mux.Router, ctxRoot string, staticPath string) {

	defaultFile := filepath.Join(staticPath, "index.html")

	fileServer := func(w http.ResponseWriter, r *http.Request) {

		log := app_http.ExtractLogger(r.Context(), "ui")

		requestedFile := filepath.Join(staticPath, strings.TrimPrefix(r.URL.Path, ctxRoot+"/ui"))

		requestedFileStats, err := os.Stat(requestedFile)
		if os.IsNotExist(err) {
			requestedFile = defaultFile
			requestedFileStats, _ = os.Stat(requestedFile)
		}

		session := app_http.ExtractSession(r.Context())
		log.Trace().
			Bool("in_context", session != nil).
			Msg("Session found")

		test, found := session.Values["test"]
		if !found {
			aRandom, _ := uuid.NewRandom()
			session.Values["test"] = aRandom.String()
		}
		log.Trace().
			Any("initial_value", test).
			Any("current_value", session.Values["test"]).
			Bool("found", found).
			Msg("Session value set")

		session.Save(r, w)

		if requestedFileStats.IsDir() {
			http.ServeFile(w, r, defaultFile)
		} else {
			http.ServeFile(w, r, requestedFile)
		}
	}

	staticRouter.Methods(http.MethodGet).HandlerFunc(fileServer).Name(ctxRoot + "/ui")
}
