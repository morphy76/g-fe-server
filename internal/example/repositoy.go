package example

import (
	"g-fe-server/pkg/example"
)

type MemoryRepository struct {
	db map[int]example.Example
}

func (r *MemoryRepository) FindAll() ([]example.Example, error) {
	values := make([]example.Example, 0, len(r.db))
	for _, v := range r.db {
		values = append(values, v)
	}
	return values, nil
}

func (r *MemoryRepository) FindById(id int) (example.Example, error) {
	return example.Example{}, nil
}

func (r *MemoryRepository) Save(e example.Example) error {
	return nil
}

func (r *MemoryRepository) Update(e example.Example) error {
	return nil
}

func (r *MemoryRepository) Delete(id int) error {
	return nil
}
