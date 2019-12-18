package mongo

import (
	"fmt"
	. "github.com/parsable/go-scim/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2"
	"gopkg.in/ory-am/dockertest.v3"
	"log"
	"os"
	"testing"
)

var (
	testSession    *mgo.Session
	dbName         string = "test"
	collectionName string = "user"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("mongo", "3.0", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		testSession, err = mgo.Dial(fmt.Sprintf("localhost:%s", resource.GetPort("27017/tcp")))
		if err != nil {
			return err
		}

		return testSession.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()
	pool.Purge(resource)
	os.Exit(code)
}

func TestRepository_Create(t *testing.T) {
	defer cleanUp()

	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)

	r, _, err := ParseResource("../resources/tests/user_1.json")
	require.Nil(t, err)

	repo := getTestRepository(sch)
	repo.Create(r, ctx)

	count, err := testSession.Copy().DB(dbName).C(collectionName).Count()
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestRepository_Get(t *testing.T) {
	defer cleanUp()

	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)

	r, _, err := ParseResource("../resources/tests/user_1.json")
	require.Nil(t, err)

	testSession.Copy().DB(dbName).C(collectionName).Insert(r.Complex)

	repo := getTestRepository(sch)

	for _, test := range []struct {
		id        string
		version   string
		assertion func(provider DataProvider, err error)
	}{
		{
			r.GetId(),
			"",
			func(provider DataProvider, err error) {
				assert.NotNil(t, provider)
				assert.Nil(t, err)
				assert.Equal(t, r.GetId(), provider.GetId())
			},
		},
		{
			r.GetId(),
			"W/\"a330bc54f0671c9\"",
			func(provider DataProvider, err error) {
				assert.NotNil(t, provider)
				assert.Nil(t, err)
				assert.Equal(t, r.GetId(), provider.GetId())
			},
		},
		{
			"foo",
			"",
			func(provider DataProvider, err error) {
				assert.Nil(t, provider)
				assert.NotNil(t, err)
				assert.IsType(t, &ResourceNotFoundError{}, err)
				assert.Equal(t, "foo", err.(*ResourceNotFoundError).Id)
			},
		},
		{
			r.GetId(),
			"invalidVersion",
			func(provider DataProvider, err error) {
				assert.Nil(t, provider)
				assert.NotNil(t, err)
				assert.IsType(t, &ResourceNotFoundError{}, err)
			},
		},
	} {
		test.assertion(repo.Get(test.id, test.version, ctx))
	}
}

func TestRepository_Count(t *testing.T) {
	defer cleanUp()

	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)

	r, _, err := ParseResource("../resources/tests/user_1.json")
	require.Nil(t, err)

	repo := getTestRepository(sch)

	count, err := repo.Count("id pr", ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)

	testSession.Copy().DB(dbName).C(collectionName).Insert(r.Complex)
	count, err = repo.Count("id pr", ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestRepository_Update(t *testing.T) {
	defer cleanUp()

	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)

	r, _, err := ParseResource("../resources/tests/user_1.json")
	require.Nil(t, err)

	testSession.Copy().DB(dbName).C(collectionName).Insert(r.Complex)

	repo := getTestRepository(sch)

	r.Complex["userName"] = "foo"
	version := r.GetData()["meta"].(map[string]interface{})["version"].(string)
	repo.Update(r.GetId(), version, r, ctx)

	r0, err := repo.Get(r.GetId(), "", ctx)
	assert.Nil(t, err)
	assert.Equal(t, r.GetId(), r0.GetId())
	assert.Equal(t, "foo", r0.GetData()["userName"])
}

func TestRepository_Delete(t *testing.T) {
	defer cleanUp()

	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)

	r, _, err := ParseResource("../resources/tests/user_1.json")
	require.Nil(t, err)

	repo := getTestRepository(sch)
	err = repo.Delete(r.GetId(), "", ctx)
	assert.NotNil(t, err)
	assert.IsType(t, &ResourceNotFoundError{}, err)

	testSession.Copy().DB(dbName).C(collectionName).Insert(r.Complex)
	err = repo.Delete(r.GetId(), "", ctx)
	assert.Nil(t, err)

	count, err := testSession.Copy().DB(dbName).C(collectionName).Count()
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
}

func TestRepository_Search(t *testing.T) {
	defer cleanUp()

	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)

	for _, name := range []string{
		"user_1", "anne", "jack", "linda", "mary", "mike", "tom",
	} {
		r, _, err := ParseResource(fmt.Sprintf("../resources/tests/%s.json", name))
		require.Nil(t, err)
		testSession.Copy().DB(dbName).C(collectionName).Insert(r.Complex)
	}

	repo := getTestRepository(sch)
	for _, test := range []struct {
		payload   SearchRequest
		assertion func(response *ListResponse, err error)
	}{
		{
			SearchRequest{
				Filter:     "id pr",
				Count:      10,
				StartIndex: 1,
			},
			func(response *ListResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 7, len(response.Resources))
			},
		},
		{
			SearchRequest{
				Filter:     "id pr",
				Count:      2,
				StartIndex: 1,
			},
			func(response *ListResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 2, len(response.Resources))
			},
		},
		{
			SearchRequest{
				Filter:     "userName sw \"david\" or userName eq \"anne\"",
				Count:      10,
				StartIndex: 1,
			},
			func(response *ListResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 2, len(response.Resources))
			},
		},
	} {
		test.assertion(repo.Search(test.payload, ctx))
	}
}

func cleanUp() {
	testSession.Copy().DB(dbName).C(collectionName).RemoveAll(nil)
}

func getTestRepository(sch *Schema) Repository {
	return &repository{
		schema:     sch,
		db:         dbName,
		collection: collectionName,
		session:    testSession,
		constructor: func(c Complex) DataProvider {
			return &Resource{Complex: c}
		},
	}
}
