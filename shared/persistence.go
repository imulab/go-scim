package shared

type DataProvider interface {
	GetId() string
	GetData() Complex
}

type Repository interface {
	Create(provider DataProvider) error

	Get(id string) (DataProvider, error)

	GetAll() ([]Complex, error)

	Count(query string) (int, error)

	Update(provider DataProvider) error

	Delete(id string) error

	Search(payload SearchRequest) (ListResponse, error)
}
