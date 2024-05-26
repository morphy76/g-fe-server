package static

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/options"
)

func HandleStatic(staticRouter *mux.Router, context context.Context) {

	ctxRoot := context.Value(app_http.CTX_CONTEXT_SERVE_KEY).(*options.ServeOptions).ContextRoot
	staticPath := context.Value(app_http.CTX_CONTEXT_SERVE_KEY).(*options.ServeOptions).StaticPath

	defaultFile := filepath.Join(staticPath, "index.html")

	fileServer := func(w http.ResponseWriter, r *http.Request) {

		path := filepath.Join(staticPath, strings.TrimPrefix(r.URL.Path, ctxRoot+"/ui"))

		fi, err := os.Stat(path)

		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		session := r.Context().Value(app_http.CTX_SESSION_KEY).(*sessions.Session)
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

		if fi.IsDir() {
			http.ServeFile(w, r, defaultFile)
		} else {
			http.ServeFile(w, r, path)
		}
	}

	staticRouter.Methods(http.MethodGet).HandlerFunc(fileServer).Name(ctxRoot + "/ui")
}
