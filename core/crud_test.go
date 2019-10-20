package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCrudGet(t *testing.T) {
	tests := []struct {
		name   string
		prop   Crud
		step   *step
		expect func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "get simple value from complex property",
			prop: Properties.NewComplexOf(
				&Attribute{Type: TypeComplex, SubAttributes: []*Attribute{
					{Name: "userName", Type: TypeString},
				}},
				map[string]interface{}{
					"username": "foo",
				},
			),
			step: Steps.NewPath("userName"),
			expect: func(t *testing.T, v interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foo", v)
			},
		},
		{
			name: "get simple value from complex property",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name: "name",
							Type: TypeComplex,
							SubAttributes: []*Attribute{
								{Name: "firstName", Type: TypeString},
							},
						},
					},
				},
				map[string]interface{}{
					"name": map[string]interface{}{
						"firstName": "foo",
					},
				},
			),
			step: Steps.NewPathChain("name", "firstName"),
			expect: func(t *testing.T, v interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foo", v)
			},
		},
		{
			name: "get sub property value from multiValued complex property",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{Name: "value", Type: TypeString},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{"value": "foo@bar.com"},
						map[string]interface{}{"value": "bar@foo.com"},
					},
				},
			),
			step: Steps.NewPathChain("emails", "value"),
			expect: func(t *testing.T, v interface{}, err error) {
				assert.Nil(t, err)
				assert.Len(t, v, 2)
				assert.Contains(t, v, "foo@bar.com")
				assert.Contains(t, v, "bar@foo.com")
			},
		},
		{
			name: "get filtered complex property",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{Name: "value", Type: TypeString},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{"value": "foo@bar.com"},
						map[string]interface{}{"value": "bar@foo.com"},
					},
				},
			),
			// emails[value eq "foo@bar.com"]
			step: &step{
				Token: "emails",
				Typ:   stepPath,
				Next:  &step{
					Token: Eq,
					Typ:   stepRelationalOperator,
					Left:  &step{
						Token: "value",
						Typ:   stepPath,
					},
					Right: &step{
						Token: "foo@bar.com",
						Typ:   stepLiteral,
					},
				},
			},
			expect: func(t *testing.T, v interface{}, err error) {
				assert.Nil(t, err)
				assert.Len(t, v, 1)
				assert.Equal(t, "foo@bar.com", v.([]interface{})[0].(map[string]interface{})["value"])
			},
		},
		{
			name: "get sub property from filtered complex property",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{Name: "value", Type: TypeString},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{"value": "foo@bar.com"},
						map[string]interface{}{"value": "bar@foo.com"},
					},
				},
			),
			// emails[value eq "foo@bar.com"].value
			step: &step{
				Token: "emails",
				Typ:   stepPath,
				Next:  &step{
					Token: Eq,
					Typ:   stepRelationalOperator,
					Left:  &step{
						Token: "value",
						Typ:   stepPath,
					},
					Right: &step{
						Token: "foo@bar.com",
						Typ:   stepLiteral,
					},
					Next: &step{
						Token: "value",
						Typ:   stepPath,
					},
				},
			},
			expect: func(t *testing.T, v interface{}, err error) {
				assert.Nil(t, err)
				assert.Len(t, v, 1)
				assert.Equal(t, "foo@bar.com", v.([]interface{})[0])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := test.prop.Get(test.step)
			test.expect(t, v, err)
		})
	}
}
