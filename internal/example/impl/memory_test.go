package example

import (
	"testing"

	"github.com/morphy76/g-fe-server/pkg/example"
)

func TestMemoryRepositorySuite(t *testing.T) {
	t.Log("Test MemoryRepository Suite")

	repo := NewMemoryRepository()
	t.Logf("Repository URL: memory")

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
