package db

import (
	"context"
	"reflect"
	"time"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/options"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateHealthCheck(dbOptions *options.DbOptions) app_http.HealthCheckFn {
	return testDbStatus
}

func testDbStatus(requestContext context.Context) (string, app_http.Status) {

	dbStatus := app_http.Inactive
	label := ""

	timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbOptions := ExtractDbOptions(requestContext)
	if dbOptions.Type == options.RepositoryTypeMemoryDB {
		dbStatus = app_http.Active
		label = "MemoryDB"
	} else if dbOptions.Type == options.RepositoryTypeMongoDB {
		dbClient := ExtractDb(requestContext)
		label = "MongoDB"
		if reflect.TypeOf(dbClient) == reflect.TypeOf(&mongo.Client{}) {

			mongoClient := dbClient.(*mongo.Client)
			errChan := make(chan error, 1)

			go func() {
				errChan <- mongoClient.Ping(timeoutContext, nil)
			}()

			select {
			case <-timeoutContext.Done():
				dbStatus = app_http.Inactive
			case err := <-errChan:
				if err != nil {
					dbStatus = app_http.Inactive
				} else {
					dbStatus = app_http.Active
				}
			}
		}
	}

	return label, dbStatus
}
