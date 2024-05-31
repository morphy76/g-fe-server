package api

import (
	"net/http"

	"github.com/morphy76/g-fe-server/internal/example/repository"
	"github.com/rs/zerolog"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	model "github.com/morphy76/g-fe-server/pkg/example"
)

type ContextualizedApiHandler func(zerolog.Logger, model.Repository) http.HandlerFunc

func ContextualizedApi(apiHandler ContextualizedApiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog := app_http.ExtractLogger(r.Context(), "example")
		exampleRepository, err := repository.NewRepository(r.Context())
		if err != nil {
			useLog.Error().
				Err(err).
				Msg("Failed to create repository")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		apiHandler(useLog, exampleRepository)(w, r)
	}
}
