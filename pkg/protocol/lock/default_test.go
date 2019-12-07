package lock

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

func TestDefaultLock(t *testing.T) {
	s := new(DefaultLockTestSuite)
	s.resourceBase = "../../tests/default_lock_test_suite"
	suite.Run(t, s)
}

type DefaultLockTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *DefaultLockTestSuite) TestLockTimeout() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	resource := prop.NewResourceOf(resourceType, map[string]interface{}{
		"id": "12B3657A-0951-4821-8386-315CF7EBC394",
	})

	lock := Default()
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	err := lock.Lock(ctx, resource)
	assert.Nil(s.T(), err)

	err = lock.Lock(ctx, resource)
	assert.Equal(s.T(), context.DeadlineExceeded, err)
}

func (s *DefaultLockTestSuite) TestConcurrentLock() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	resource := prop.NewResourceOf(resourceType, map[string]interface{}{
		"id": "12B3657A-0951-4821-8386-315CF7EBC394",
	})

	lock := Default()
	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func(id int, l Lock) {
			defer wg.Done()
			err := lock.Lock(context.Background(), resource)
			s.T().Logf("%d acquired lock", id)
			assert.Nil(s.T(), err)
			lock.Unlock(context.Background(), resource)
		}(i, lock)
	}
	wg.Wait()
}

func (s *DefaultLockTestSuite) TestSequentialLock() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	resource := prop.NewResourceOf(resourceType, map[string]interface{}{
		"id": "12B3657A-0951-4821-8386-315CF7EBC394",
	})

	lock := Default()

	for i := 0; i < 100; i++ {
		err := lock.Lock(context.Background(), resource)
		assert.Nil(s.T(), err)
		lock.Unlock(context.Background(), resource)
	}
}

func (s *DefaultLockTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *DefaultLockTestSuite) mustSchema(filePath string) *spec.Schema {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	sch := new(spec.Schema)
	err = json.Unmarshal(raw, sch)
	s.Require().Nil(err)

	spec.SchemaHub.Put(sch)

	return sch
}
