package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name   string
		prop   Evaluation
		root   *Step
		expect func(t *testing.T, r bool, err error)
	}{
		{
			name: "eq predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "userName", Type: TypeString},
					},
				},
				map[string]interface{}{
					"userName": "foo",
				},
			),
			root: &Step{
				Token: Eq,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "userName",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "foo",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "ne predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "userName", Type: TypeString},
					},
				},
				map[string]interface{}{
					"userName": "foo",
				},
			),
			root: &Step{
				Token: Ne,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "userName",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "bar",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "gt predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "age", Type: TypeInteger},
					},
				},
				map[string]interface{}{
					"age": int64(10),
				},
			),
			root: &Step{
				Token: Gt,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "age",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "8",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "lt predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "created", Type: TypeDateTime},
					},
				},
				map[string]interface{}{
					"created": "2019-10-10T10:10:10",
				},
			),
			root: &Step{
				Token: Lt,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "created",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "2019-11-11T11:11:11",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "ge predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "age", Type: TypeInteger},
					},
				},
				map[string]interface{}{
					"age": int64(10),
				},
			),
			root: &Step{
				Token: Ge,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "age",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "10",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "le predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "created", Type: TypeDateTime},
					},
				},
				map[string]interface{}{
					"created": "2019-10-10T10:10:10",
				},
			),
			root: &Step{
				Token: Le,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "created",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "2019-10-10T10:10:10",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "pr predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeString,
					SubAttributes: []*Attribute{
						{Name: "schemas", Type: TypeString},
					},
				},
				map[string]interface{}{
					"schemas": []interface{}{},
				},
			),
			root: &Step{
				Token: Pr,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "schemas",
					Typ:   stepPath,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.False(t, r)
			},
		},
		{
			name: "sw predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "userName", Type: TypeString},
					},
				},
				map[string]interface{}{
					"userName": "abc",
				},
			),
			root: &Step{
				Token: Sw,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "userName",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "a",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "ew predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "userName", Type: TypeString},
					},
				},
				map[string]interface{}{
					"userName": "abc",
				},
			),
			root: &Step{
				Token: Ew,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "userName",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "c",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "co predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "userName", Type: TypeString},
					},
				},
				map[string]interface{}{
					"userName": "abc",
				},
			),
			root: &Step{
				Token: Co,
				Typ:   stepRelationalOperator,
				Left: &Step{
					Token: "userName",
					Typ:   stepPath,
				},
				Right: &Step{
					Token: "b",
					Typ:   stepLiteral,
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "and predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "userName", Type: TypeString},
						{Name: "age", Type: TypeInteger},
					},
				},
				map[string]interface{}{
					"userName": "foo",
					"age":      int64(10),
				},
			),
			root: &Step{
				Token: And,
				Typ:   stepLogicalOperator,
				Left: &Step{
					Token: Eq,
					Typ:   stepRelationalOperator,
					Left: &Step{
						Token: "userName",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "foo",
						Typ:   stepLiteral,
					},
				},
				Right: &Step{
					Token: Gt,
					Typ:   stepRelationalOperator,
					Left: &Step{
						Token: "age",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "8",
						Typ:   stepLiteral,
					},
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "or predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "userName", Type: TypeString},
						{Name: "age", Type: TypeInteger},
					},
				},
				map[string]interface{}{
					"userName": "foo",
					"age":      int64(10),
				},
			),
			root: &Step{
				Token: Or,
				Typ:   stepLogicalOperator,
				Left: &Step{
					Token: Ne,
					Typ:   stepRelationalOperator,
					Left: &Step{
						Token: "userName",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "foo",
						Typ:   stepLiteral,
					},
				},
				Right: &Step{
					Token: Gt,
					Typ:   stepRelationalOperator,
					Left: &Step{
						Token: "age",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "8",
						Typ:   stepLiteral,
					},
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
		{
			name: "not predicate",
			prop: Properties.NewComplexOf(
				&Attribute{
					Type: TypeComplex,
					SubAttributes: []*Attribute{
						{Name: "userName", Type: TypeString},
					},
				},
				map[string]interface{}{
					"userName": "foo",
				},
			),
			root: &Step{
				Token: Not,
				Typ:   stepLogicalOperator,
				Left: &Step{
					Token: Eq,
					Typ:   stepRelationalOperator,
					Left: &Step{
						Token: "userName",
						Typ:   stepPath,
					},
					Right: &Step{
						Token: "bar",
						Typ:   stepLiteral,
					},
				},
			},
			expect: func(t *testing.T, r bool, err error) {
				assert.Nil(t, err)
				assert.True(t, r)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := test.prop.Evaluate(test.root)
			test.expect(t, r, err)
		})
	}
}
