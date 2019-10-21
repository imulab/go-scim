package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsEqualTo(t *testing.T) {
	tests := []struct {
		name   string
		prop   EqualAware
		value  interface{}
		expect bool
	}{
		{
			name:   "equal string",
			prop:   Properties.NewStringOf(&Attribute{Name: "test", Type: TypeString, CaseExact: true}, "foo"),
			value:  "foo",
			expect: true,
		},
		{
			name:   "unequal string",
			prop:   Properties.NewStringOf(&Attribute{Name: "test", Type: TypeString, CaseExact: true}, "foo"),
			value:  "bar",
			expect: false,
		},
		{
			name:   "equal string (case insensitive)",
			prop:   Properties.NewStringOf(&Attribute{Name: "test", Type: TypeString, CaseExact: false}, "foo"),
			value:  "FOO",
			expect: true,
		},
		{
			name: "equal integer",
			prop: Properties.NewIntegerOf(&Attribute{Name: "test", Type: TypeInteger}, 100),
			value: 100,
			expect: true,
		},
		{
			name: "unequal integer",
			prop: Properties.NewIntegerOf(&Attribute{Name: "test", Type: TypeInteger}, 100),
			value: 200,
			expect: false,
		},
		{
			name: "equal decimal",
			prop: Properties.NewDecimalOf(&Attribute{Name: "test", Type: TypeDecimal}, 100.123),
			value: 100.123,
			expect: true,
		},
		{
			name: "unequal decimal",
			prop: Properties.NewDecimalOf(&Attribute{Name: "test", Type: TypeDecimal}, 100.123),
			value: 200.123,
			expect: false,
		},
		{
			name: "equal boolean",
			prop: Properties.NewBooleanOf(&Attribute{Name: "test", Type: TypeBoolean}, true),
			value: true,
			expect: true,
		},
		{
			name: "unequal boolean",
			prop: Properties.NewBooleanOf(&Attribute{Name: "test", Type: TypeBoolean}, true),
			value: false,
			expect: false,
		},
		{
			name: "equal boolean (nil is treated as false)",
			prop: Properties.NewBoolean(&Attribute{Name: "test", Type: TypeBoolean}),
			value: false,
			expect: true,
		},
		{
			name: "equal dateTime",
			prop: Properties.NewDateTimeOf(&Attribute{Name: "test", Type: TypeDateTime}, "2019-01-01T10:10:10"),
			value: "2019-01-01T10:10:10",
			expect: true,
		},
		{
			name: "unequal dateTime",
			prop: Properties.NewDateTimeOf(&Attribute{Name: "test", Type: TypeDateTime}, "2019-01-01T10:10:10"),
			value: "2009-01-01T10:10:10",
			expect: false,
		},
		{
			name: "equal binary",
			prop: Properties.NewBinaryOf(&Attribute{Name: "test", Type: TypeBinary}, "Zm9vCg=="),
			value: "Zm9vCg==",
			expect: true,
		},
		{
			name: "equal binary",
			prop: Properties.NewBinaryOf(&Attribute{Name: "test", Type: TypeBinary}, "Zm9vCg=="),
			value: "YmFyCg==",
			expect: false,
		},
		{
			name: "equal reference",
			prop: Properties.NewReferenceOf(&Attribute{Name: "test", Type: TypeReference}, "Users/foo"),
			value: "Users/foo",
			expect: true,
		},
		{
			name: "unequal reference",
			prop: Properties.NewReferenceOf(&Attribute{Name: "test", Type: TypeReference}, "Users/foo"),
			value: "Users/bar",
			expect: false,
		},
		{
			name: "matching multiValue",
			prop: &multiValuedProperty{
				attr: &Attribute{Name:"test", Type: TypeString, MultiValued: true},
				props: []Property{
					Properties.NewStringOf(&Attribute{Name:"test", Type: TypeString}, "foo"),
					Properties.NewStringOf(&Attribute{Name:"test", Type: TypeString}, "bar"),
				},
			},
			value: "foo",
			expect: true,
		},
		{
			name: "matching multiValue",
			prop: &multiValuedProperty{
				attr: &Attribute{Name:"test", Type: TypeString, MultiValued: true},
				props: []Property{
					Properties.NewStringOf(&Attribute{Name:"test", Type: TypeString}, "foo"),
					Properties.NewStringOf(&Attribute{Name:"test", Type: TypeString}, "bar"),
				},
			},
			value: "foobar",
			expect: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eq := test.prop.IsEqualTo(test.value)
			assert.Equal(t, test.expect, eq)
		})
	}
}
