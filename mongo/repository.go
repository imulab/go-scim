package mongo

import (
	"fmt"
	. "github.com/davidiamyou/go-scim/shared"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func NewMongoRepositoryWithUrl(url, db, collection string, sch *Schema, constructor func(Complex) DataProvider) (Repository, error) {
	repo := &repository{}
	if session, err := mgo.Dial(url); err != nil {
		return nil, err
	} else {
		repo.session = session
		repo.db = db
		repo.collection = collection
		repo.schema = sch
		repo.constructor = constructor
		if err := repo.ensureIndexes(); err != nil {
			return nil, err
		}
		return repo, nil
	}
}

func NewMongoRepository(
	info *mgo.DialInfo,
	db, collection string,
	sch *Schema,
	constructor func(Complex) DataProvider,
) (Repository, error) {
	repo := &repository{dialInfo: info}
	if session, err := mgo.DialWithInfo(repo.dialInfo); err != nil {
		return nil, err
	} else {
		repo.session = session
		repo.db = db
		repo.collection = collection
		repo.schema = sch
		repo.constructor = constructor
		if err := repo.ensureIndexes(); err != nil {
			return nil, err
		}
		return repo, nil
	}
}

type repository struct {
	db          string
	collection  string
	schema      *Schema
	constructor func(Complex) DataProvider
	dialInfo    *mgo.DialInfo
	session     *mgo.Session
}

func (r *repository) getCollection() (*mgo.Collection, func()) {
	s := r.session.Copy()
	return s.DB(r.db).C(r.collection), func() { s.Close() }
}

func (r *repository) construct(c Complex) DataProvider {
	if r.constructor != nil {
		return r.constructor(c)
	} else {
		return &Resource{Complex: c}
	}
}

func (r *repository) ensureIndexes() error {
	return nil
}

func (r *repository) handleError(err error, args ...interface{}) error {
	if err == nil {
		return nil
	}
	switch {
	case err.Error() == "not found":
		if len(args) > 1 {
			return Error.ResourceNotFound(
				fmt.Sprintf("%v", args[0]),
				fmt.Sprintf("%v", args[1]),
			)
		} else if len(args) > 0 {
			return Error.ResourceNotFound(
				fmt.Sprintf("%v", args[0]),
				"",
			)
		}
		return Error.ResourceNotFound("", "")
	default:
		return err
	}
}

func (r *repository) Create(provider DataProvider) error {
	c, cleanUp := r.getCollection()
	defer cleanUp()

	return r.handleError(c.Insert(provider.GetData()))
}

func (r *repository) Get(id, version string) (DataProvider, error) {
	c, cleanUp := r.getCollection()
	defer cleanUp()

	data := make(map[string]interface{}, 0)
	var query bson.M
	if len(version) == 0 {
		query = bson.M{"id": id}
	} else {
		query = bson.M{"id": id, "meta.version": version}
	}
	err := c.Find(query).One(&data)
	if err != nil {
		return nil, r.handleError(err, id)
	}

	delete(data, "_id")
	return r.construct(Complex(data)), nil
}

func (r *repository) GetAll() ([]Complex, error) {
	panic("not supported")
}

func (r *repository) Count(query string) (int, error) {
	q, err := convertToMongoQuery(query, r.schema)
	if err != nil {
		return 0, r.handleError(err)
	}

	c, cleanUp := r.getCollection()
	defer cleanUp()

	count, err := c.Find(q).Count()
	return count, r.handleError(err)
}

func (r *repository) Update(id, version string, provider DataProvider) error {
	c, cleanUp := r.getCollection()
	defer cleanUp()

	var query bson.M
	if len(version) == 0 {
		query = bson.M{"id": id}
	} else {
		query = bson.M{"id": id, "meta.version": version}
	}
	err := c.Update(query, provider.GetData())
	return r.handleError(err, provider.GetId())
}

func (r *repository) Delete(id, version string) error {
	c, cleanUp := r.getCollection()
	defer cleanUp()

	var query bson.M
	if len(version) == 0 {
		query = bson.M{"id": id}
	} else {
		query = bson.M{"id": id, "meta.version": version}
	}
	err := c.Remove(query)
	return r.handleError(err, id)
}

func (r *repository) Search(payload SearchRequest) (*ListResponse, error) {
	c, cleanUp := r.getCollection()
	defer cleanUp()

	q, err := convertToMongoQuery(payload.Filter, r.schema)
	if err != nil {
		return nil, r.handleError(err)
	}

	totalResults, err := c.Find(q).Count()
	if err != nil {
		return nil, r.handleError(err)
	}

	query := c.Find(q)
	if len(payload.SortBy) > 0 {
		if payload.Ascending() {
			query = query.Sort(payload.SortBy)
		} else {
			query = query.Sort("-" + payload.SortBy)
		}
	}
	query = query.Skip(payload.StartIndex - 1)
	query = query.Limit(payload.Count)

	listData := make([]map[string]interface{}, 0)
	err = query.Iter().All(&listData)
	if err != nil {
		return nil, r.handleError(err)
	}

	results := make([]DataProvider, 0, len(listData))
	for _, elem := range listData {
		results = append(results, r.construct(Complex(elem)))
	}

	return &ListResponse{
		Schemas:      []string{ListResponseUrn},
		StartIndex:   payload.StartIndex,
		ItemsPerPage: payload.Count,
		TotalResults: totalResults,
		Resources:    results,
	}, nil
}
