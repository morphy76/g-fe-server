package db

import (
	"context"
	"time"

	"github.com/morphy76/g-fe-server/cmd/options"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CreateHealthCheck creates a health check function for the db
func CreateHealthCheck(dbOptions *options.MongoDBOptions) app_http.AdditionalCheckFn {
	client, err := NewClient(dbOptions)
	if err != nil {
		panic(err)
	}
	return testDbStatus(client)
}

func testDbStatus(client *mongo.Client) app_http.AdditionalCheckFn {
	return func(requestContext context.Context) (app_http.HealthCheckFn, app_http.Probe) {
		return func(requestContext context.Context) (string, app_http.Status) {

			dbStatus := app_http.Inactive
			label := "MongoDB"

			timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			errChan := make(chan error, 1)

			go func() {
				errChan <- client.Ping(timeoutContext, nil)
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

			return label, dbStatus
		}, app_http.Live | app_http.Ready
	}
}
