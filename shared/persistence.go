package shared

type DataProvider interface {
	GetId() string
	GetData() Complex
}

type Repository interface {
	Create(provider DataProvider) error

	Get(id, version string) (DataProvider, error)

	GetAll() ([]Complex, error)

	Count(query string) (int, error)

	Update(id, version string, provider DataProvider) error

	Delete(id, version string) error

	Search(payload SearchRequest) (*ListResponse, error)
}

// An simple in memory database fit for test use and read only production use
// this implementation:
// - only implements Create, Get, GetAll, Update, Delete
// - is not thread safe
// - ignores the version argument
type mapRepository struct {
	data map[string]DataProvider
}

func (r *mapRepository) Create(provider DataProvider) error {
	r.data[provider.GetId()] = provider
	return nil
}

func (r *mapRepository) Get(id, version string) (DataProvider, error) {
	if dp, ok := r.data[id]; !ok {
		return nil, Error.ResourceNotFound(id, version)
	} else {
		return dp, nil
	}
}

func (r *mapRepository) GetAll() ([]Complex, error) {
	all := make([]Complex, 0)
	for _, v := range r.data {
		all = append(all, v.GetData())
	}
	return all, nil
}

func (r *mapRepository) Count(query string) (int, error) {
	return 0, Error.Text("not implemented")
}

func (r *mapRepository) Update(id, version string, provider DataProvider) error {
	if _, ok := r.data[id]; !ok {
		return Error.ResourceNotFound(id, version)
	} else {
		r.data[id] = provider
		return nil
	}
}

func (r *mapRepository) Delete(id, version string) error {
	if _, ok := r.data[id]; !ok {
		return Error.ResourceNotFound(id, version)
	} else {
		delete(r.data, id)
		return nil
	}
}

func (r *mapRepository) Search(payload SearchRequest) (*ListResponse, error) {
	return nil, Error.Text("not implemented")
}

func NewMapRepository(initialData map[string]DataProvider) Repository {
	if len(initialData) == 0 {
		return &mapRepository{data: make(map[string]DataProvider, 0)}
	} else {
		return &mapRepository{data: initialData}
	}
}
