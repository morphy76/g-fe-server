package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func HandleOpenAPI(
	parent *mux.Router,
	ctxRoot string,
) {
	healthRouter := parent.PathPrefix("/openapi").Subrouter()
	healthRouter.Methods(http.MethodGet).HandlerFunc(onOpenAPI).Name("GET " + ctxRoot + "/api/openapi")
}

func onOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./api/openapi.json")
}
