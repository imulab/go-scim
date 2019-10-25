package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStartsWith(t *testing.T) {
	tests := []struct {
		name   string
		prop   StringOpCapable
		value  interface{}
		expect bool
	}{
		{
			name:   "foo starts with f (case sensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "foo"),
			value:  "f",
			expect: true,
		},
		{
			name:   "foo starts with F (case insensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: false}, "foo"),
			value:  "F",
			expect: true,
		},
		{
			name:   "foo does not start with F (case sensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "foo"),
			value:  "F",
			expect: false,
		},
		{
			name:   "reference User starts with U (case should always be sensitive)",
			prop:   Properties.NewReferenceOf(&Attribute{Type: TypeString, CaseExact: true}, "User"),
			value:  "U",
			expect: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.prop.StartsWith(test.value)
			assert.Equal(t, test.expect, result)
		})
	}
}

func TestEndsWith(t *testing.T) {
	tests := []struct {
		name   string
		prop   StringOpCapable
		value  interface{}
		expect bool
	}{
		{
			name:   "foo ends with o (case sensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "foo"),
			value:  "o",
			expect: true,
		},
		{
			name:   "foo ends with O (case insensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: false}, "foo"),
			value:  "O",
			expect: true,
		},
		{
			name:   "foo does not end with O (case sensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "foo"),
			value:  "O",
			expect: false,
		},
		{
			name:   "reference User ends with r (case should always be sensitive)",
			prop:   Properties.NewReferenceOf(&Attribute{Type: TypeString, CaseExact: true}, "User"),
			value:  "r",
			expect: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.prop.EndsWith(test.value)
			assert.Equal(t, test.expect, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		prop   StringOpCapable
		value  interface{}
		expect bool
	}{
		{
			name:   "foo contains o (case sensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "foo"),
			value:  "o",
			expect: true,
		},
		{
			name:   "foo contains O (case insensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: false}, "foo"),
			value:  "O",
			expect: true,
		},
		{
			name:   "foo does not contain O (case sensitive)",
			prop:   Properties.NewStringOf(&Attribute{Type: TypeString, CaseExact: true}, "foo"),
			value:  "O",
			expect: false,
		},
		{
			name:   "reference User contains se (case should always be sensitive)",
			prop:   Properties.NewReferenceOf(&Attribute{Type: TypeString, CaseExact: true}, "User"),
			value:  "se",
			expect: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.prop.Contains(test.value)
			assert.Equal(t, test.expect, result)
		})
	}
}
