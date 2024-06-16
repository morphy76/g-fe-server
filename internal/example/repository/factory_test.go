package repository

import (
	"context"
	"testing"

	"github.com/morphy76/g-fe-server/internal/db"
	example "github.com/morphy76/g-fe-server/internal/example/repository/impl"
	"github.com/morphy76/g-fe-server/internal/options"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestFactorySuite(t *testing.T) {
	t.Log("Test Factory Suite")

	testContext := db.InjectDb(context.Background(), &mongo.Client{})

	t.Run("Test RepositoryTypeMemoryDB", func(t *testing.T) {
		t.Log("Test Factory RepositoryTypeMemoryDB")
		useContext := db.InjectDbOptions(testContext, &options.DbOptions{
			Type: options.RepositoryTypeMemoryDB,
		})
		if repo, err := NewRepository(useContext); err != nil {
			t.Fatalf("Failed to create the repository: %s", err)
		} else if _, ok := repo.(*example.MemoryRepository); !ok {
			t.Fatalf("Expected Repository got %T", repo)
		}
	})

	t.Run("Test RepositoryTypeMongoDB", func(t *testing.T) {
		t.Log("Test Factory RepositoryTypeMongoDB")
		useContext := db.InjectDbOptions(testContext, &options.DbOptions{
			Type: options.RepositoryTypeMongoDB,
		})
		if repo, err := NewRepository(useContext); err != nil {
			t.Fatalf("Failed to create the repository: %s", err)
		} else if _, ok := repo.(*example.MongoRepository); !ok {
			t.Fatalf("Expected Repository got %T", repo)
		}
	})
}
