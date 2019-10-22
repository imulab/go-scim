package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCrudGet(t *testing.T) {
	tests := []struct {
		name   string
		prop   Crud
		step   *Step
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
			step: &Step{
				Token: "emails",
				Typ:   stepPath,
				Next: &Step{
					Token: Eq,
					Typ:   stepRelationalOperator,
					Left: &Step{
						Token: "value",
						Typ:   stepPath,
					},
					Right: &Step{
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
			step: &Step{
				Token: "emails",
				Typ:   stepPath,
				Next: &Step{
					Token: Eq,
					Typ:   stepRelationalOperator,
					Left: &Step{
						Token: "value",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "foo@bar.com",
						Typ:   stepLiteral,
					},
					Next: &Step{
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

func TestCrudAdd(t *testing.T) {
	tests := []struct {
		name   string
		prop   Crud
		step   *Step
		value  interface{}
		expect func(t *testing.T, prop Crud, err error)
	}{
		{
			name: "add value to unassigned simple property",
			prop: Properties.NewComplex(
				&Attribute{Type: TypeComplex, SubAttributes: []*Attribute{
					{Name: "userName", Type: TypeString},
				}}),
			step:  Steps.NewPath("userName"),
			value: "foo",
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("userName"))
				assert.Equal(t, "foo", v)
			},
		},
		{
			name: "add value to assigned simple property",
			prop: Properties.NewComplexOf(
				&Attribute{Type: TypeComplex, SubAttributes: []*Attribute{
					{Name: "userName", Type: TypeString},
				}},
				map[string]interface{}{
					"username": "foo",
				},
			),
			step:  Steps.NewPath("userName"),
			value: "bar",
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("userName"))
				assert.Equal(t, "bar", v)
			},
		},
		{
			name: "add value to nested simple property",
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
			step:  Steps.NewPathChain("name", "firstName"),
			value: "bar",
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPathChain("name", "firstName"))
				assert.Equal(t, "bar", v)
			},
		},
		{
			name: "add value to multiValued property",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "tags", Type: TypeString, MultiValued: true},
					},
				},
				map[string]interface{}{
					"tags": []interface{}{"one", "two"},
				},
			),
			step:  Steps.NewPath("tags"),
			value: "three",
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("tags"))
				assert.Len(t, v, 3)
				assert.Contains(t, v, "one")
				assert.Contains(t, v, "two")
				assert.Contains(t, v, "three")
			},
		},
		{
			name: "add value to a multiValued complex property (and switch exclusive)",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{
									Name:     "value",
									Type:     TypeString,
									Metadata: &Metadata{IsIdentity: true},
								},
								{
									Name:     "primary",
									Type:     TypeBoolean,
									Metadata: &Metadata{IsIdentity: true, IsExclusive: true},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{
							"value":   "foo@bar.com",
							"primary": true,
						},
						map[string]interface{}{
							"value": "foo2@bar.com",
						},
					},
				},
			),
			step:   Steps.NewPath("emails"),
			value:  map[string]interface{}{
				"value": "foo3@bar.com",
				"primary": true,
			},
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("emails"))
				assert.Len(t, v, 3)

				getPrimaryOfValue := func(value string) interface{} {
					v, err := prop.Get(&Step{
						Token: "emails",
						Typ:   stepPath,
						Next:  &Step{
							Token: Eq,
							Typ:   stepRelationalOperator,
							Left:  &Step{
								Token: "value",
								Typ:   stepPath,
							},
							Right: &Step{
								Token: value,
								Typ:   stepLiteral,
							},
							Next: &Step{
								Token: "primary",
								Typ:   stepPath,
							},
						},
					})
					assert.Nil(t, err)
					return v
				}

				assert.Len(t, getPrimaryOfValue("foo@bar.com"), 0)
				assert.Len(t, getPrimaryOfValue("foo2@bar.com"), 0)
				assert.Len(t, getPrimaryOfValue("foo3@bar.com"), 1)
			},
		},
		{
			name: "add value to a sub property of a multiValued complex property (and switch exclusive)",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{
									Name:     "value",
									Type:     TypeString,
									Metadata: &Metadata{IsIdentity: true},
								},
								{
									Name:     "primary",
									Type:     TypeBoolean,
									Metadata: &Metadata{IsIdentity: true, IsExclusive: true},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{
							"value":   "foo@bar.com",
							"primary": true,
						},
						map[string]interface{}{
							"value": "foo2@bar.com",
						},
					},
				},
			),
			step:   &Step{
				Token: "emails",
				Typ:   stepPath,
				Next:  &Step{
					Token: Eq,
					Typ:   stepRelationalOperator,
					Left:  &Step{
						Token: "value",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "foo2@bar.com",
						Typ:   stepLiteral,
					},
					Next: &Step{
						Token: "primary",
						Typ:   stepPath,
					},
				},
			},
			value:  true,
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("emails"))
				assert.Len(t, v, 2)

				getPrimaryOfValue := func(value string) interface{} {
					v, err := prop.Get(&Step{
						Token: "emails",
						Typ:   stepPath,
						Next:  &Step{
							Token: Eq,
							Typ:   stepRelationalOperator,
							Left:  &Step{
								Token: "value",
								Typ:   stepPath,
							},
							Right: &Step{
								Token: value,
								Typ:   stepLiteral,
							},
							Next: &Step{
								Token: "primary",
								Typ:   stepPath,
							},
						},
					})
					assert.Nil(t, err)
					return v
				}

				assert.Len(t, getPrimaryOfValue("foo@bar.com"), 0)
				assert.Len(t, getPrimaryOfValue("foo2@bar.com"), 1)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.prop.Add(test.step, test.value)
			test.expect(t, test.prop, err)
		})
	}
}

func TestCrudReplace(t *testing.T) {
	tests := []struct {
		name   string
		prop   Crud
		step   *Step
		value  interface{}
		expect func(t *testing.T, prop Crud, err error)
	}{
		{
			name: "replace value of unassigned simple property",
			prop: Properties.NewComplex(
				&Attribute{Type: TypeComplex, SubAttributes: []*Attribute{
					{Name: "userName", Type: TypeString},
				}}),
			step:  Steps.NewPath("userName"),
			value: "foo",
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("userName"))
				assert.Equal(t, "foo", v)
			},
		},
		{
			name: "replace value of assigned simple property",
			prop: Properties.NewComplexOf(
				&Attribute{Type: TypeComplex, SubAttributes: []*Attribute{
					{Name: "userName", Type: TypeString},
				}},
				map[string]interface{}{
					"username": "foo",
				},
			),
			step:  Steps.NewPath("userName"),
			value: "bar",
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("userName"))
				assert.Equal(t, "bar", v)
			},
		},
		{
			name: "replace value of nested simple property",
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
			step:  Steps.NewPathChain("name", "firstName"),
			value: "bar",
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPathChain("name", "firstName"))
				assert.Equal(t, "bar", v)
			},
		},
		{
			name: "replace element of a multiValued complex property (and switch exclusive)",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{
									Name:     "value",
									Type:     TypeString,
									Metadata: &Metadata{IsIdentity: true},
								},
								{
									Name:     "primary",
									Type:     TypeBoolean,
									Metadata: &Metadata{IsIdentity: true, IsExclusive: true},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{
							"value":   "foo@bar.com",
							"primary": true,
						},
						map[string]interface{}{
							"value": "foo2@bar.com",
						},
					},
				},
			),
			step:   &Step{
				Token: "emails",
				Typ:   stepPath,
				Next:  &Step{
					Token: Ne,
					Typ:   stepRelationalOperator,
					Left:  &Step{
						Token: "primary",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "true",
						Typ:   stepLiteral,
					},
				},
			},
			value:  map[string]interface{}{
				"value": "foo3@bar.com",
				"primary": true,
			},
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("emails"))
				assert.Len(t, v, 2)

				getPrimaryOfValue := func(value string) interface{} {
					v, err := prop.Get(&Step{
						Token: "emails",
						Typ:   stepPath,
						Next:  &Step{
							Token: Eq,
							Typ:   stepRelationalOperator,
							Left:  &Step{
								Token: "value",
								Typ:   stepPath,
							},
							Right: &Step{
								Token: value,
								Typ:   stepLiteral,
							},
							Next: &Step{
								Token: "primary",
								Typ:   stepPath,
							},
						},
					})
					assert.Nil(t, err)
					return v
				}

				assert.Len(t, getPrimaryOfValue("foo@bar.com"), 0)
				assert.Len(t, getPrimaryOfValue("foo3@bar.com"), 1)
			},
		},
		{
			name: "replace sub property of an element in a multiValued complex property (and switch exclusive)",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{
									Name:     "value",
									Type:     TypeString,
									Metadata: &Metadata{IsIdentity: true},
								},
								{
									Name:     "primary",
									Type:     TypeBoolean,
									Metadata: &Metadata{IsIdentity: true, IsExclusive: true},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{
							"value":   "foo@bar.com",
							"primary": true,
						},
						map[string]interface{}{
							"value": "foo2@bar.com",
						},
					},
				},
			),
			step:   &Step{
				Token: "emails",
				Typ:   stepPath,
				Next:  &Step{
					Token: Ne,
					Typ:   stepRelationalOperator,
					Left:  &Step{
						Token: "primary",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "true",
						Typ:   stepLiteral,
					},
					Next: &Step{
						Token: "primary",
						Typ:   stepPath,
					},
				},
			},
			value:  true,
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("emails"))
				assert.Len(t, v, 2)

				getPrimaryOfValue := func(value string) interface{} {
					v, err := prop.Get(&Step{
						Token: "emails",
						Typ:   stepPath,
						Next:  &Step{
							Token: Eq,
							Typ:   stepRelationalOperator,
							Left:  &Step{
								Token: "value",
								Typ:   stepPath,
							},
							Right: &Step{
								Token: value,
								Typ:   stepLiteral,
							},
							Next: &Step{
								Token: "primary",
								Typ:   stepPath,
							},
						},
					})
					assert.Nil(t, err)
					return v
				}

				assert.Len(t, getPrimaryOfValue("foo@bar.com"), 0)
				assert.Len(t, getPrimaryOfValue("foo2@bar.com"), 1)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.prop.Replace(test.step, test.value)
			test.expect(t, test.prop, err)
		})
	}
}

func TestCrudDelete(t *testing.T) {
	tests := []struct {
		name   string
		prop   Crud
		step   *Step
		expect func(t *testing.T, prop Crud, err error)
	}{
		{
			name: "delete value from assigned simple property",
			prop: Properties.NewComplexOf(
				&Attribute{Type: TypeComplex, SubAttributes: []*Attribute{
					{Name: "userName", Type: TypeString},
				}},
				map[string]interface{}{
					"username": "foo",
				},
			),
			step:  Steps.NewPath("userName"),
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, err := prop.Get(Steps.NewPath("userName"))
				assert.Nil(t, err)
				assert.Nil(t, v)
			},
		},
		{
			name: "delete value from nested simple property",
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
			step:  Steps.NewPathChain("name", "firstName"),
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, err := prop.Get(Steps.NewPathChain("name", "firstName"))
				assert.Nil(t, err)
				assert.Nil(t, v)
			},
		},
		{
			name: "delete element from a multiValued complex property",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{
									Name:     "value",
									Type:     TypeString,
									Metadata: &Metadata{IsIdentity: true},
								},
								{
									Name:     "primary",
									Type:     TypeBoolean,
									Metadata: &Metadata{IsIdentity: true, IsExclusive: true},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{
							"value":   "foo@bar.com",
							"primary": true,
						},
						map[string]interface{}{
							"value": "foo2@bar.com",
						},
					},
				},
			),
			step:   &Step{
				Token: "emails",
				Typ:   stepPath,
				Next:  &Step{
					Token: Ne,
					Typ:   stepRelationalOperator,
					Left:  &Step{
						Token: "primary",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "true",
						Typ:   stepLiteral,
					},
				},
			},
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("emails"))
				assert.Len(t, v, 1)
				assert.Equal(t, "foo@bar.com", v.([]interface{})[0].(map[string]interface{})["value"])
			},
		},
		{
			name: "delete sub property of an element in a multiValued complex property",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{
							Name:        "emails",
							Type:        TypeComplex,
							MultiValued: true,
							SubAttributes: []*Attribute{
								{
									Name:     "value",
									Type:     TypeString,
									Metadata: &Metadata{IsIdentity: true},
								},
								{
									Name:     "primary",
									Type:     TypeBoolean,
									Metadata: &Metadata{IsIdentity: true, IsExclusive: true},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{
							"value":   "foo@bar.com",
							"primary": true,
						},
						map[string]interface{}{
							"value": "foo2@bar.com",
						},
					},
				},
			),
			step:   &Step{
				Token: "emails",
				Typ:   stepPath,
				Next:  &Step{
					Token: Eq,
					Typ:   stepRelationalOperator,
					Left:  &Step{
						Token: "primary",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "true",
						Typ:   stepLiteral,
					},
					Next: &Step{
						Token: "primary",
						Typ:   stepPath,
					},
				},
			},
			expect: func(t *testing.T, prop Crud, err error) {
				assert.Nil(t, err)
				v, _ := prop.Get(Steps.NewPath("emails"))
				assert.Len(t, v, 2)

				getPrimaryOfValue := func(value string) interface{} {
					v, err := prop.Get(&Step{
						Token: "emails",
						Typ:   stepPath,
						Next:  &Step{
							Token: Eq,
							Typ:   stepRelationalOperator,
							Left:  &Step{
								Token: "value",
								Typ:   stepPath,
							},
							Right: &Step{
								Token: value,
								Typ:   stepLiteral,
							},
							Next: &Step{
								Token: "primary",
								Typ:   stepPath,
							},
						},
					})
					assert.Nil(t, err)
					return v
				}

				assert.Len(t, getPrimaryOfValue("foo@bar.com"), 0)
				assert.Len(t, getPrimaryOfValue("foo2@bar.com"), 0)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.prop.Delete(test.step)
			test.expect(t, test.prop, err)
		})
	}
}
