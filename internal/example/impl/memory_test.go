package example

import (
	"testing"

	"github.com/morphy76/g-fe-server/pkg/example"

	"context"
)

func MemoryCleanup(repo example.Repository, t *testing.T, ctx context.Context) func() {
	return func() {
		t.Log("Cleanup Repository Suite")

		if err := repo.Disconnect(); err != nil {
			t.Logf("Could not disconnect the repository: %s", err)
		} else {
			t.Log("Repository disconnected")
		}
	}
}

func TestMemoryRepositorySuite(t *testing.T) {
	t.Log("Test MemoryRepository Suite")

	ctx := context.Background()

	repo := NewMemoryRepository()
	t.Logf("Repository URI: memory")

	if err := repo.Connect(); err != nil {
		t.Fatalf("Failed to connect the repository: %s", err)
	} else {
		t.Log("Repository connected")
	}

	t.Cleanup(MemoryCleanup(repo, t, ctx))

	t.Run("Test List", func(t *testing.T) {
		t.Log("Testing Memory List")

		repo.Save(example.Example{
			Name: "Test",
			Age:  10,
		})
		repo.Save(example.Example{
			Name: "Test2",
			Age:  20,
		})

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
