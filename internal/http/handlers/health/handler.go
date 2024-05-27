package health

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/options"
)

func HealthHandlers(nonFunctionalRouter *mux.Router, ctxRoot string, dbOptions *options.DbOptions) {

	healthRouter := nonFunctionalRouter.Path("/health").Subrouter()
	healthRouter.Use(middleware.JSONResponse)

	healthRouter.Methods(http.MethodGet).HandlerFunc(onHealth()).Name(ctxRoot + "/g/health")
}

func onHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		overallStatus := Active

		timeoutContext, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		label, dbStatus := testDbStatus(timeoutContext)
		if dbStatus == Inactive {
			overallStatus = Inactive
		}

		if overallStatus == Active {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(&HealthResponse{
			Status: overallStatus,
			SubSystems: map[string]HealthResponse{
				label: {Status: dbStatus},
			},
		})
	}
}

func testDbStatus(requestContext context.Context) (string, Status) {

	dbStatus := Inactive
	label := ""

	timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbOptions := requestContext.Value(app_http.CTX_DB_OPTIONS_KEY).(*options.DbOptions)
	if dbOptions.Type == options.RepositoryTypeMemoryDB {
		dbStatus = Active
		label = "MemoryDB"
	} else if dbOptions.Type == options.RepositoryTypeMongoDB {
		dbClient := requestContext.Value(app_http.CTX_DB_KEY)
		label = "MongoDB"
		if reflect.TypeOf(dbClient) == reflect.TypeOf(&mongo.Client{}) {

			mongoClient := dbClient.(*mongo.Client)
			errChan := make(chan error, 1)

			go func() {
				errChan <- mongoClient.Ping(timeoutContext, nil)
			}()

			select {
			case <-timeoutContext.Done():
				dbStatus = Inactive
			case err := <-errChan:
				if err != nil {
					dbStatus = Inactive
				} else {
					dbStatus = Active
				}
			}
		}
	}

	return label, dbStatus
}
