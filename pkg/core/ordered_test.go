package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsGreaterThan(t *testing.T) {
	tests := []struct {
		name   string
		prop   OrderAware
		value  interface{}
		expect bool
	}{
		{
			name:   "z is greater than a (case exact)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "z"),
			value:  "a",
			expect: true,
		},
		{
			name:   "a is greater than Z (case exact)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "a"),
			value:  "Z",
			expect: true,
		},
		{
			name:   "a is not greater than Z (not case exact)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: false}, "a"),
			value:  "Z",
			expect: false,
		},
		{
			name:   "200 is greater than 100",
			prop:   Properties.NewIntegerOf(&Attribute{Type: TypeInteger}, 200),
			value:  100,
			expect: true,
		},
		{
			name:   "200 is not greater than 200",
			prop:   Properties.NewIntegerOf(&Attribute{Type: TypeInteger}, 200),
			value:  200,
			expect: false,
		},
		{
			name:   "200.123 is greater than 100.123",
			prop:   Properties.NewDecimalOf(&Attribute{Type: TypeDecimal}, 200.123),
			value:  100.123,
			expect: true,
		},
		{
			name:   "200.123 is not greater than 200.123",
			prop:   Properties.NewDecimalOf(&Attribute{Type: TypeDecimal}, 200.123),
			value:  200.123,
			expect: false,
		},
		{
			name:   "year 2019 is greater than year 2018",
			prop:   Properties.NewDateTimeOf(&Attribute{Type: TypeDateTime}, "2019-10-10T10:10:10"),
			value:  "2018-10-10T10:10:10",
			expect: true,
		},
		{
			name:   "year 2019 is not greater than year 2020",
			prop:   Properties.NewDateTimeOf(&Attribute{Type: TypeDateTime}, "2019-10-10T10:10:10"),
			value:  "2020-10-10T10:10:10",
			expect: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.prop.IsGreaterThan(test.value)
			assert.Equal(t, test.expect, result)
		})
	}
}

func TestIsLessThan(t *testing.T) {
	tests := []struct {
		name   string
		prop   OrderAware
		value  interface{}
		expect bool
	}{
		{
			name:   "a is less than z (case exact)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "a"),
			value:  "z",
			expect: true,
		},
		{
			name:   "Z is less than a (case exact)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "Z"),
			value:  "a",
			expect: true,
		},
		{
			name:   "Z is not less than a (not case exact)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: false}, "Z"),
			value:  "a",
			expect: false,
		},
		{
			name:   "100 is less than 200",
			prop:   Properties.NewIntegerOf(&Attribute{Type: TypeInteger}, 100),
			value:  200,
			expect: true,
		},
		{
			name:   "200 is not less than 200",
			prop:   Properties.NewIntegerOf(&Attribute{Type: TypeInteger}, 200),
			value:  200,
			expect: false,
		},
		{
			name:   "100.123 is less than 200.123",
			prop:   Properties.NewDecimalOf(&Attribute{Type: TypeDecimal}, 100.123),
			value:  200.123,
			expect: true,
		},
		{
			name:   "200.123 is not less than 200.123",
			prop:   Properties.NewDecimalOf(&Attribute{Type: TypeDecimal}, 200.123),
			value:  200.123,
			expect: false,
		},
		{
			name:   "year 2018 is less than year 2019",
			prop:   Properties.NewDateTimeOf(&Attribute{Type: TypeDateTime}, "2018-10-10T10:10:10"),
			value:  "2019-10-10T10:10:10",
			expect: true,
		},
		{
			name:   "year 2020 is not less than year 2019",
			prop:   Properties.NewDateTimeOf(&Attribute{Type: TypeDateTime}, "2020-10-10T10:10:10"),
			value:  "2019-10-10T10:10:10",
			expect: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.prop.IsLessThan(test.value)
			assert.Equal(t, test.expect, result)
		})
	}
}
