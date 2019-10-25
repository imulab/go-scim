package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPropertyRaw(t *testing.T) {
	var (
		str = "foo"
		i   = int64(100)
		d   = 1.23
		b   = true
		dt  = "2019-10-17T10:10:10Z"
		ref = "User"
	)

	tests := []struct {
		name     string
		property Property
		expect   func(t *testing.T, v interface{})
	}{
		{
			name:     "unassigned string reports nil value",
			property: &stringProperty{},
			expect: func(t *testing.T, v interface{}) {
				assert.Nil(t, v)
			},
		},
		{
			name:     "assigned string reports string value",
			property: &stringProperty{v: &str},
			expect: func(t *testing.T, v interface{}) {
				assert.Equal(t, str, v)
			},
		},
		{
			name:     "unassigned integer reports nil value",
			property: &integerProperty{},
			expect: func(t *testing.T, v interface{}) {
				assert.Nil(t, v)
			},
		},
		{
			name:     "assigned integer reports int64 value",
			property: &integerProperty{v: &i},
			expect: func(t *testing.T, v interface{}) {
				assert.Equal(t, i, v)
			},
		},
		{
			name:     "unassigned decimal reports nil value",
			property: &decimalProperty{},
			expect: func(t *testing.T, v interface{}) {
				assert.Nil(t, v)
			},
		},
		{
			name:     "assigned decimal reports float64 value",
			property: &decimalProperty{v: &d},
			expect: func(t *testing.T, v interface{}) {
				assert.Equal(t, d, v)
			},
		},
		{
			name:     "unassigned boolean reports nil value",
			property: &booleanProperty{},
			expect: func(t *testing.T, v interface{}) {
				assert.Nil(t, v)
			},
		},
		{
			name:     "assigned boolean reports bool value",
			property: &booleanProperty{v: &b},
			expect: func(t *testing.T, v interface{}) {
				assert.Equal(t, b, v)
			},
		},
		{
			name:     "unassigned dateTime reports nil value",
			property: &dateTimeProperty{},
			expect: func(t *testing.T, v interface{}) {
				assert.Nil(t, v)
			},
		},
		{
			name:     "assigned dateTime reports string value",
			property: &dateTimeProperty{v: &dt},
			expect: func(t *testing.T, v interface{}) {
				assert.Equal(t, dt, v)
			},
		},
		{
			name:     "unassigned reference reports nil value",
			property: &referenceProperty{},
			expect: func(t *testing.T, v interface{}) {
				assert.Nil(t, v)
			},
		},
		{
			name:     "assigned reference reports string value",
			property: &referenceProperty{v: &ref},
			expect: func(t *testing.T, v interface{}) {
				assert.Equal(t, ref, v)
			},
		},
		{
			name:     "unassigned binary reports nil value",
			property: &binaryProperty{},
			expect: func(t *testing.T, v interface{}) {
				assert.Nil(t, v)
			},
		},
		{
			name:     "assigned binary reports string value",
			property: &binaryProperty{v: &str},
			expect: func(t *testing.T, v interface{}) {
				assert.Equal(t, str, v)
			},
		},
		{
			name:     "unassigned multiValued reports zero length slice",
			property: &multiValuedProperty{},
			expect: func(t *testing.T, v interface{}) {
				assert.Len(t, v, 0)
			},
		},
		{
			name: "complex reports empty map",
			property: &complexProperty{
				subProps: map[string]Property{
					"foo": &stringProperty{v: &str, attr: &Attribute{Name: "foo"}},
					"bar": &stringProperty{attr: &Attribute{Name: "bar"}},
				},
			},
			expect: func(t *testing.T, v interface{}) {
				assert.IsType(t, map[string]interface{}{}, v)
				assert.Equal(t, "foo", v.(map[string]interface{})["foo"])
				assert.Nil(t, v.(map[string]interface{})["bar"])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.expect(t, test.property.Raw())
		})
	}
}

func TestPropertyUnassigned(t *testing.T) {
	var (
		str = "foo"
		i   = int64(100)
		d   = 1.23
		b   = true
		dt  = "2019-10-17T10:10:10Z"
		ref = "User"
	)

	tests := []struct {
		name     string
		property Property
		expect   bool
	}{
		{
			name:     "nil string is unassigned",
			property: &stringProperty{},
			expect:   true,
		},
		{
			name:     "nil integer is unassigned",
			property: &integerProperty{},
			expect:   true,
		},
		{
			name:     "nil decimal is unassigned",
			property: &decimalProperty{},
			expect:   true,
		},
		{
			name:     "nil boolean is unassigned",
			property: &booleanProperty{},
			expect:   true,
		},
		{
			name:     "nil dateTime is unassigned",
			property: &dateTimeProperty{},
			expect:   true,
		},
		{
			name:     "nil reference is unassigned",
			property: &referenceProperty{},
			expect:   true,
		},
		{
			name:     "nil binary is unassigned",
			property: &binaryProperty{},
			expect:   true,
		},
		{
			name:     "nil complex is unassigned",
			property: &complexProperty{},
			expect:   true,
		},
		{
			name:     "nil multiValued is unassigned",
			property: &multiValuedProperty{},
			expect:   true,
		},
		{
			name: "complex containing all unassigned property is unassigned",
			property: &complexProperty{
				subProps: map[string]Property{
					"foo": &stringProperty{},
					"bar": &booleanProperty{},
				},
			},
			expect: true,
		},
		{
			name: "empty multiValued is unassigned",
			property: &multiValuedProperty{
				props: []Property{},
			},
			expect: true,
		},
		{
			name:     "non-nil string is assigned",
			property: &stringProperty{v: &str},
			expect:   false,
		},
		{
			name:     "non-nil integer is assigned",
			property: &integerProperty{v: &i},
			expect:   false,
		},
		{
			name:     "non-nil decimal is assigned",
			property: &decimalProperty{v: &d},
			expect:   false,
		},
		{
			name:     "non-nil boolean is assigned",
			property: &booleanProperty{v: &b},
			expect:   false,
		},
		{
			name:     "non-nil dateTime is assigned",
			property: &dateTimeProperty{v: &dt},
			expect:   false,
		},
		{
			name:     "non-nil reference is assigned",
			property: &referenceProperty{v: &ref},
			expect:   false,
		},
		{
			name:     "non-nil binary is assigned",
			property: &binaryProperty{v: &str},
			expect:   false,
		},
		{
			name: "non-empty complex is assigned",
			property: &complexProperty{
				subProps: map[string]Property{
					"foo": &stringProperty{v: &str},
					"bar": &stringProperty{},
				},
			},
			expect: false,
		},
		{
			name:     "non-empty multiValued is assigned",
			property: &multiValuedProperty{props: []Property{&stringProperty{}}},
			expect:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.property.IsUnassigned())
		})
	}
}
