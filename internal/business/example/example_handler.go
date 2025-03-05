package example

import (
	"net/http"

	"github.com/Unleash/unleash-client-go/v4"
	featContext "github.com/Unleash/unleash-client-go/v4/context"
	"github.com/gorilla/mux"
	"github.com/morphy76/g-fe-server/internal/common"
	"github.com/morphy76/g-fe-server/internal/features"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
	"github.com/rs/zerolog"
)

const (
	moduleName            = "module.example"
	modulePath            = "/example"
	failedToWriteResponse = "Failed to write response"
)

func Handler(
	apiRouter *mux.Router,
	feServer *server.FEServer,
	routerLog zerolog.Logger,
) {
	exampleRouter := apiRouter.PathPrefix(modulePath).Subrouter()
	featCtx := featContext.Context{
		Properties: map[string]string{
			"role": "api",
		},
	}
	exampleRouter.Use(features.ModuleVerifier(moduleName, unleash.WithContext(featCtx)))
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Str("module", moduleName).
			Msg("Router registered")
	}

	exampleRouter.HandleFunc("/up", doUpHandler).Name("GET " + feServer.ServeOpts.ContextRoot + "/api/" + moduleName + "/up")
	exampleRouter.HandleFunc("/down", doDownHandler).Name("GET " + feServer.ServeOpts.ContextRoot + "/api/" + moduleName + "/down")

	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Str("module", moduleName).
			Msg("Handler registered")
	}
}

func doUpHandler(w http.ResponseWriter, r *http.Request) {
	useLogger := logger.GetLogger(r.Context(), "http."+moduleName)

	service := NewExampleService(r.Context())

	service.DoUp()

	downAnswer, err := service.CallDown()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err = common.ToJSON(NewExampleResponse(err.Error()), w)
		if err != nil {
			useLogger.Error().Err(err).Msg(failedToWriteResponse)
		}
		return
	}

	err = common.ToJSON(NewExampleResponse("From AIW: "+downAnswer.Message), w)
	if err != nil {
		useLogger.Error().Err(err).Msg(failedToWriteResponse)
	}
}

func doDownHandler(w http.ResponseWriter, r *http.Request) {
	useLogger := logger.GetLogger(r.Context(), "http."+moduleName)
	service := NewExampleService(r.Context())

	rv, err := service.DoDown()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err = common.ToJSON(NewExampleResponse(err.Error()), w)
		if err != nil {
			useLogger.Error().Err(err).Msg(failedToWriteResponse)
		}
		return
	}

	err = common.ToJSON(NewExampleResponse(rv), w)
	if err != nil {
		useLogger.Error().Err(err).Msg(failedToWriteResponse)
	}
}
