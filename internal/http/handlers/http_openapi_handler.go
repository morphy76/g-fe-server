package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// HandleOpenAPI registers the OpenAPI handler for the given parent router.
func HandleOpenAPI(
	parent *mux.Router,
	ctxRoot string,
) error {
	healthRouter := parent.PathPrefix("/openapi").Subrouter()
	healthRouter.Methods(http.MethodGet).HandlerFunc(onOpenAPI).Name("GET " + ctxRoot + "/api/openapi")
	return nil
}

func onOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./api/openapi.json")
}
