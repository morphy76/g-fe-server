package example

type Example struct {
	Name string `json:"name" db:"name" validate:"required"`
	Age  int    `json:"age" db:"age" validate:"required"`
}

type RepositoryType int8

const (
	RepositoryTypeMemoryDB RepositoryType = iota
	RepositoryTypeMongoDB  RepositoryType = 1
)

type Repository interface {
	FindAll() ([]Example, error)
	FindById(id string) (Example, error)
	Save(e Example) error
	Update(e Example) error
	Delete(id string) error
	Connect() error
	Disconnect() error
	IsConnected() bool
	Ping() bool
}
