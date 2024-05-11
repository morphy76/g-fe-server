package example

type Repository interface {
	FindAll() ([]Example, error)
	FindById(id int) (Example, error)
	Save(e Example) error
	Update(e Example) error
	Delete(id int) error
}
