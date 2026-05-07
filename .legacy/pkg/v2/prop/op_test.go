package prop

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type OperatorTestSuite struct{}

func (s *OperatorTestSuite) testEqualTo(t *testing.T, prop Property, v interface{}, expect bool) {
	if eqProp, ok := prop.(EqCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, eqProp.EqualsTo(v))
	}
}

func (s *OperatorTestSuite) testStartsWith(t *testing.T, prop Property, v string, expect bool) {
	if swProp, ok := prop.(SwCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, swProp.StartsWith(v))
	}
}

func (s *OperatorTestSuite) testEndsWith(t *testing.T, prop Property, v string, expect bool) {
	if ewProp, ok := prop.(EwCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, ewProp.EndsWith(v))
	}
}

func (s *OperatorTestSuite) testContains(t *testing.T, prop Property, v string, expect bool) {
	if coProp, ok := prop.(CoCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, coProp.Contains(v))
	}
}

func (s *OperatorTestSuite) testGreaterThan(t *testing.T, prop Property, v interface{}, expect bool) {
	if gtProp, ok := prop.(GtCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, gtProp.GreaterThan(v))
	}
}

func (s *OperatorTestSuite) testGreaterThanOrEqualTo(t *testing.T, prop Property, v interface{}, expect bool) {
	if geProp, ok := prop.(GeCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, geProp.GreaterThanOrEqualTo(v))
	}
}

func (s *OperatorTestSuite) testLessThan(t *testing.T, prop Property, v interface{}, expect bool) {
	if ltProp, ok := prop.(LtCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, ltProp.LessThan(v))
	}
}

func (s *OperatorTestSuite) testLessThanOrEqualTo(t *testing.T, prop Property, v interface{}, expect bool) {
	if leProp, ok := prop.(LeCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, leProp.LessThanOrEqualTo(v))
	}
}

func (s *OperatorTestSuite) testPresent(t *testing.T, prop Property, expect bool) {
	if prProp, ok := prop.(PrCapable); !ok {
		assert.False(t, expect)
	} else {
		assert.Equal(t, expect, prProp.Present())
	}
}
