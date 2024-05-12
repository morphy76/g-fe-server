package example

import (
	"g-fe-server/pkg/example"
)

type MemoryRepository struct {
	db map[string]example.Example
}

func NewMemoryRepository() example.Repository {
	return &MemoryRepository{
		db: make(map[string]example.Example),
	}
}

func (r *MemoryRepository) FindAll() ([]example.Example, error) {
	values := make([]example.Example, 0, len(r.db))
	for _, v := range r.db {
		values = append(values, v)
	}
	return values, nil
}

func (r *MemoryRepository) FindById(id string) (example.Example, error) {

	rv, ok := r.db[id]
	if !ok {
		return rv, example.ErrNotFound
	}
	return rv, nil
}

func (r *MemoryRepository) Save(e example.Example) error {

	if _, ok := r.db[e.Name]; ok {
		return example.ErrAlreadyExists
	}
	r.db[e.Name] = e
	return nil
}

func (r *MemoryRepository) Update(e example.Example) error {
	appo, ok := r.db[e.Name]
	if !ok {
		return example.ErrNotFound
	}
	appo.Age = e.Age
	r.db[e.Name] = appo
	return nil
}

func (r *MemoryRepository) Delete(id string) error {
	if _, ok := r.db[id]; !ok {
		return example.ErrNotFound
	}
	delete(r.db, id)
	return nil
}

func (r *MemoryRepository) Connect() error {
	return nil
}

func (r *MemoryRepository) Disconnect() error {
	return nil
}

func (r *MemoryRepository) IsConnected() bool {
	return true
}

func (r *MemoryRepository) Ping() bool {
	return true
}
