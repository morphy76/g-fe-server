package example

import (
	"fmt"
	"testing"

	"context"

	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

const db_name = "go_db"

func MongoCleanup(mongoC *mongodb.MongoDBContainer, t *testing.T, ctx context.Context) func() {
	return func() {
		t.Log("Cleanup Repository Suite")

		if err := mongoC.Terminate(ctx); err != nil {
			t.Logf("Could not stop MongoDB: %s", err)
		} else {
			t.Log("MongoDB container stopped")
		}
	}
}

func TestMongoRepositorySuite(t *testing.T) {
	t.Log("Test MongoRepository Suite")

	ctx := context.Background()

	mongoC, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:7"),
		testcontainers.WithEnv(map[string]string{
			"MONGO_INITDB_DATABASE":      db_name,
			"MONGO_INITDB_ROOT_USERNAME": "go_root",
			"MONGO_INITDB_ROOT_PASSWORD": "go_password",
		}),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Files: []testcontainers.ContainerFile{
					{
						HostFilePath:      "./test_resources/init.js",
						ContainerFilePath: "/docker-entrypoint-initdb.d/init.js",
						FileMode:          0644,
					},
				},
			},
		}),
	)
	if err != nil {
		t.Fatalf("Failed to start container: %s", err)
	} else {
		t.Log("MongoDB container started")
	}

	host, err := mongoC.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get host: %s", err)
	}

	ports, err := mongoC.Ports(ctx)
	if err != nil {
		t.Fatalf("Failed to get ports: %s", err)
	}

	usePort := ports["27017/tcp"][0]

	t.Cleanup(MongoCleanup(mongoC, t, ctx))

	dbOptions := &options.DbOptions{
		Type: options.RepositoryTypeMongoDB,
		MongoDbOptions: options.MongoDbOptions{
			Url:      fmt.Sprintf("mongodb://%s:%s/%s", host, usePort.HostPort, db_name),
			User:     "go",
			Password: "go",
		},
	}
	dbClient, err := db.NewClient(dbOptions)
	if err != nil {
		t.Fatalf("Failed to create the client: %s", err)
	}

	repo := &MongoRepository{
		DbOptions:  dbOptions,
		Client:     dbClient.(*mongo.Client),
		UseContext: ctx,
	}
	t.Logf("Repository URL: %s", dbOptions.Url)

	t.Run("Test List", func(t *testing.T) {
		t.Log("Testing Mongo List")

		items, err := repo.FindAll()
		if err != nil {
			t.Errorf("Error on List: %s", err)
		}
		if len(items) == 0 {
			t.Error("Expected items")
		}

		for _, item := range items {
			t.Logf("%#v", item)
		}
	})
}
