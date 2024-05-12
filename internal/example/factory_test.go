package example

import (
	"testing"

	impl "g-fe-server/internal/example/impl"
	app_context "g-fe-server/internal/http/context"
)

func TestFactorySuite(t *testing.T) {
	t.Log("Test Factory Suite")

	t.Run("Test RepositoryTypeMemoryDB", func(t *testing.T) {
		t.Log("Test Factory RepositoryTypeMemoryDB")
		if repo, err := NewRepository(app_context.DbModel{
			Type: RepositoryTypeMemoryDB,
		}); err != nil {
			t.Fatalf("Failed to create the repository: %s", err)
		} else if _, ok := repo.(*impl.MemoryRepository); !ok {
			t.Fatalf("Expected Repository got %T", repo)
		}
	})

	t.Run("Test RepositoryTypeMongoDB", func(t *testing.T) {
		t.Log("Test Factory RepositoryTypeMongoDB")
		if repo, err := NewRepository(app_context.DbModel{
			Type: RepositoryTypeMemoryDB,
		}); err != nil {
			t.Fatalf("Failed to create the repository: %s", err)
		} else if _, ok := repo.(*impl.MongoRepository); !ok {
			t.Fatalf("Expected Repository got %T", repo)
		}
	})
}
