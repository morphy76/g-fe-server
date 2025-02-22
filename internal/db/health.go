package db

import (
	"context"
	"time"

	"github.com/morphy76/g-fe-server/internal/common/health"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func CreateHealthCheck(client *mongo.Client) health.AdditionalCheckFn {
	return func(requestContext context.Context) (health.HealthCheckFn, health.Probe) {
		return func(requestContext context.Context) (string, health.Status) {

			dbStatus := health.Inactive
			label := "MongoDB"

			timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			errChan := make(chan error, 1)

			go func() {
				errChan <- client.Ping(timeoutContext, nil)
			}()

			select {
			case <-timeoutContext.Done():
				dbStatus = health.Inactive
			case err := <-errChan:
				if err != nil {
					dbStatus = health.Inactive
				} else {
					dbStatus = health.Active
				}
			}

			return label, dbStatus
		}, health.Live | health.Ready
	}
}
