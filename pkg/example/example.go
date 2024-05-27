package example

type Example struct {
	Name string `json:"name" db:"name" validate:"required"`
	Age  int    `json:"age" db:"age" validate:"required"`
}

type Repository interface {
	FindAll() ([]Example, error)
	FindById(id string) (Example, error)
	Save(e Example) error
	Update(e Example) error
	Delete(id string) error
}
