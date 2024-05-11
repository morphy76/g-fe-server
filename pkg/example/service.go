package example

type Repository interface {
	FindAll() ([]Example, error)
	FindById(id string) (Example, error)
	Save(e Example) error
	Update(e Example) error
	Delete(id string) error
}
