package example

import (
	"g-fe-server/pkg/example"
)

type memoryRepository struct {
	db map[string]example.Example
}

func NewMemoryRepository() example.Repository {
	return &memoryRepository{
		db: make(map[string]example.Example),
	}
}

func (r *memoryRepository) FindAll() ([]example.Example, error) {
	values := make([]example.Example, 0, len(r.db))
	for _, v := range r.db {
		values = append(values, v)
	}
	return values, nil
}

func (r *memoryRepository) FindById(id string) (example.Example, error) {
	return r.db[id], nil
}

func (r *memoryRepository) Save(e example.Example) error {
	r.db[e.Name] = e
	return nil
}

func (r *memoryRepository) Update(e example.Example) error {
	appo := r.db[e.Name]
	appo.Age = e.Age
	r.db[e.Name] = appo
	return nil
}

func (r *memoryRepository) Delete(id string) error {
	delete(r.db, id)
	return nil
}
