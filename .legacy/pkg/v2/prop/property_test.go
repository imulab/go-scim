package prop

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"testing"
)

// A common test suite for properties
type PropertyTestSuite struct {
	NewFunc   func(attr *spec.Attribute) Property
	NewOfFunc func(attr *spec.Attribute, v interface{}) Property
}

func (s *PropertyTestSuite) testNew(
	t *testing.T,
	before func(),
	attr *spec.Attribute,
	expect func(t *testing.T, p Property),
) {
	if before != nil {
		before()
	}
	expect(t, s.NewFunc(attr))
}

func (s *PropertyTestSuite) testRaw(
	t *testing.T,
	attr *spec.Attribute,
	value interface{},
	expect func(t *testing.T, raw interface{}),
) {
	var p Property
	if value == nil {
		p = s.NewFunc(attr)
	} else {
		p = s.NewOfFunc(attr, value)
	}
	expect(t, p.Raw())
}

func (s *PropertyTestSuite) testUnassigned(
	t *testing.T,
	attr *spec.Attribute,
	value interface{},
	expect func(t *testing.T, unassigned bool),
) {
	var p Property
	if value == nil {
		p = s.NewFunc(attr)
	} else {
		p = s.NewOfFunc(attr, value)
	}
	expect(t, p.IsUnassigned())
}

func (s *PropertyTestSuite) testDirty(
	t *testing.T,
	attr *spec.Attribute,
	mod func(t *testing.T, p Property),
	expect func(t *testing.T, dirty bool),
) {
	p := s.NewFunc(attr)
	mod(t, p)
	expect(t, p.Dirty())
}

func (s *PropertyTestSuite) testMatches(
	t *testing.T,
	a Property,
	b Property,
	expect func(t *testing.T, match bool),
) {
	expect(t, a.Matches(b))
}

func (s *PropertyTestSuite) testClone(t *testing.T, p Property) {
	p1 := p.Clone()
	assert.Equal(t, p.Attribute().ID(), p1.Attribute().ID())
	if p.IsUnassigned() {
		assert.True(t, p1.IsUnassigned())
	} else {
		assert.Equal(t, p.Raw(), p1.Raw())
	}
	assert.False(t, p == p1)
}

func (s *PropertyTestSuite) testAdd(
	t *testing.T,
	p Property,
	v interface{},
	expect func(t *testing.T, raw interface{}, err error),
) {
	_, err := p.Add(v)
	expect(t, p.Raw(), err)
}

func (s *PropertyTestSuite) testReplace(
	t *testing.T,
	p Property,
	v interface{},
	expect func(t *testing.T, raw interface{}, err error),
) {
	_, err := p.Replace(v)
	expect(t, p.Raw(), err)
}

func (s *PropertyTestSuite) testDelete(
	t *testing.T,
	p Property,
	expect func(t *testing.T, err error),
) {
	_, err := p.Delete()
	assert.True(t, p.IsUnassigned())
	expect(t, err)
}

func (s *PropertyTestSuite) mustAttribute(t *testing.T, reader io.Reader) *spec.Attribute {
	raw, err := ioutil.ReadAll(reader)
	require.Nil(t, err)

	attr := new(spec.Attribute)
	err = json.Unmarshal(raw, attr)
	require.Nil(t, err)

	return attr
}
