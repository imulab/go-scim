package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPath(t *testing.T) {
	for _, test := range []struct {
		text      string
		assertion func(head Path, err error)
	}{
		{
			// single
			"username",
			func(head Path, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, head)
				assert.Nil(t, head.Next())
				assert.Equal(t, "username", head.Base())
			},
		},
		{
			// duplex
			"name.familyName",
			func(head Path, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, head)
				assert.Equal(t, "name", head.Base())
				assert.Equal(t, "familyName", head.Next().Base())
				assert.Nil(t, head.Next().Next())
			},
		},
		{
			// single filter
			"emails[value eq \"david@foo.com\"]",
			func(head Path, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, head)
				assert.Nil(t, head.Next())
				assert.Equal(t, "emails", head.Base())
				assert.Equal(t, Eq, head.FilterRoot().Data())
			},
		},
		{
			// duplex with filter
			"emails[value eq \"david@foo.com\"].type",
			func(head Path, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, head)
				assert.Equal(t, "emails", head.Base())
				assert.Equal(t, Eq, head.FilterRoot().Data())
				assert.Equal(t, "type", head.Next().Base())
				assert.Nil(t, head.Next().Next())
			},
		},
	} {
		test.assertion(NewPath(test.text))
	}
}

func TestNewFilter(t *testing.T) {
	for _, test := range []struct {
		text      string
		assertion func(root FilterNode, err error)
	}{
		{
			// eq
			"username eq \"david\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Eq, root.Data())

				assert.NotNil(t, root.Left())
				assert.Equal(t, PathOperand, root.Left().Type())
				assert.Equal(t, "username", root.Left().Data().(Path).Base())

				assert.NotNil(t, root.Right())
				assert.Equal(t, ConstantOperand, root.Right().Type())
				assert.Equal(t, "david", root.Right().Data())
			},
		},
		{
			// ne
			"active ne true",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Ne, root.Data())

				assert.NotNil(t, root.Left())
				assert.Equal(t, PathOperand, root.Left().Type())
				assert.Equal(t, "active", root.Left().Data().(Path).Base())

				assert.NotNil(t, root.Right())
				assert.Equal(t, ConstantOperand, root.Right().Type())
				assert.Equal(t, true, root.Right().Data())
			},
		},
		{
			// sw
			"username sw \"david\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Sw, root.Data())
			},
		},
		{
			// ew
			"username ew \"david\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Ew, root.Data())
			},
		},
		{
			// co
			"username co \"david\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Co, root.Data())
			},
		},
		{
			// pr
			"emails pr",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Pr, root.Data())

				assert.NotNil(t, root.Left())
				assert.Equal(t, PathOperand, root.Left().Type())
				assert.Equal(t, "emails", root.Left().Data().(Path).Base())

				assert.Nil(t, root.Right())
			},
		},
		{
			// gt
			"created gt \"20170406\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Gt, root.Data())

				assert.NotNil(t, root.Left())
				assert.Equal(t, PathOperand, root.Left().Type())
				assert.Equal(t, "created", root.Left().Data().(Path).Base())

				assert.NotNil(t, root.Right())
				assert.Equal(t, ConstantOperand, root.Right().Type())
				assert.Equal(t, "20170406", root.Right().Data())
			},
		},
		{
			// ge
			"created ge \"20170406\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Ge, root.Data())
			},
		},
		{
			// lt
			"created lt \"20170406\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Lt, root.Data())
			},
		},
		{
			// le
			"created le \"20170406\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, RelationalOperator, root.Type())
				assert.Equal(t, Le, root.Data())
			},
		},
		{
			// and
			"username eq \"david\" and age gt 18",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, LogicalOperator, root.Type())
				assert.Equal(t, And, root.Data())

				assert.NotNil(t, root.Left())
				assert.Equal(t, RelationalOperator, root.Left().Type())
				assert.Equal(t, Eq, root.Left().Data())

				assert.NotNil(t, root.Right())
				assert.Equal(t, RelationalOperator, root.Right().Type())
				assert.Equal(t, Gt, root.Right().Data())
			},
		},
		{
			// or
			"username eq \"david\" or age gt 18",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, LogicalOperator, root.Type())
				assert.Equal(t, Or, root.Data())

				assert.NotNil(t, root.Left())
				assert.Equal(t, RelationalOperator, root.Left().Type())
				assert.Equal(t, Eq, root.Left().Data())

				assert.NotNil(t, root.Right())
				assert.Equal(t, RelationalOperator, root.Right().Type())
				assert.Equal(t, Gt, root.Right().Data())
			},
		},
		{
			// not
			"not username eq \"david\"",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, LogicalOperator, root.Type())
				assert.Equal(t, Not, root.Data())

				assert.NotNil(t, root.Left())
				assert.Equal(t, RelationalOperator, root.Left().Type())
				assert.Equal(t, Eq, root.Left().Data())

				assert.Nil(t, root.Right())
			},
		},
		{
			// parenthesis
			"not (username eq \"david\" and age gt 18)",
			func(root FilterNode, err error) {
				assert.Nil(t, err)

				assert.NotNil(t, root)
				assert.Equal(t, LogicalOperator, root.Type())
				assert.Equal(t, Not, root.Data())

				assert.NotNil(t, root.Left())
				assert.Equal(t, LogicalOperator, root.Left().Type())
				assert.Equal(t, And, root.Left().Data())

				assert.Nil(t, root.Right())
			},
		},
	} {
		test.assertion(NewFilter(test.text))
	}
}
