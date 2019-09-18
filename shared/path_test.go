package shared

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
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
		// single with UserUrn
		{
			fmt.Sprintf("%s:UserName", strings.ToLower(UserUrn)),
			func(head Path, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, head)
				assert.Equal(t, fmt.Sprintf("%s:UserName", strings.ToLower(UserUrn)), head.Base())
				assert.Nil(t, head.Next())
			},
		},
		// duplex with UserUrn
		{
			fmt.Sprintf("%s:Name.FamilyName", strings.ToLower(UserUrn)),
			func(head Path, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, head)
				assert.Equal(t, fmt.Sprintf("%s:Name", strings.ToLower(UserUrn)), head.Base())
				assert.NotNil(t, head.Next())
				assert.Equal(t, "FamilyName", head.Next().Base())
				assert.Nil(t, head.Next().Next())
			},
		},

		// single with UserEnterpriseUrn
		{
			UserEnterpriseUrn,
			func(head Path, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, head)
				assert.Equal(t, UserEnterpriseUrn, head.Base())
				assert.Nil(t, head.Next())
			},
		},
		// duplex with UserEnterpriseUrn
		{
			fmt.Sprintf("%s.department", UserEnterpriseUrn),
			func(head Path, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, head)
				assert.Equal(t, UserEnterpriseUrn, head.Base())
				assert.NotNil(t, head.Next())
				assert.Equal(t, "department", head.Next().Base())
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

func TestPath_SeparateAtLast(t *testing.T) {
	for _, test := range []struct {
		pathText  string
		assertion func(a, b Path)
	}{
		{
			// single path
			"foo",
			func(a, b Path) {
				assert.Nil(t, a)
				assert.Equal(t, "foo", b.Base())
				assert.Nil(t, b.Next())
			},
		},
		{
			// duplex path
			"name.familyName",
			func(a, b Path) {
				assert.Equal(t, "name", a.Base())
				assert.Nil(t, a.Next())
				assert.Equal(t, "familyName", b.Base())
				assert.Nil(t, b.Next())
			},
		},
		{
			// duplex path with filter
			"emails[type eq \"work\"].value",
			func(a, b Path) {
				assert.Equal(t, "emails", a.Base())
				assert.NotNil(t, a.FilterRoot())
				assert.Nil(t, a.Next())
				assert.Equal(t, "value", b.Base())
				assert.Nil(t, b.Next())
			},
		},
	} {
		p, err := NewPath(test.pathText)
		require.Nil(t, err)
		test.assertion(p.SeparateAtLast())
	}
}

func TestPath_CorrectCase(t *testing.T) {
	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)
	require.NotNil(t, sch)

	for _, test := range []struct {
		pathText  string
		assertion func(p Path)
	}{
		{
			"UserName",
			func(p Path) {
				assert.Equal(t, "userName", p.Base())
			},
		},
		{
			fmt.Sprintf("%s:UserName", strings.ToLower(UserUrn)),
			func(p Path) {
				assert.Equal(t, fmt.Sprintf("%s:userName", UserUrn), p.Base())
			},
		},
		{
			"Name.FamilyName",
			func(p Path) {
				assert.Equal(t, "name", p.Base())
				assert.Equal(t, "familyName", p.Next().Base())
			},
		},
		{
			fmt.Sprintf("%s:Name.FamilyName", strings.ToLower(UserUrn)),
			func(p Path) {
				assert.Equal(t, fmt.Sprintf("%s:name", UserUrn), p.Base())
				assert.Equal(t, "familyName", p.Next().Base())
			},
		},
		{
			"Emails[Type eq \"home\"].Value",
			func(p Path) {
				assert.Equal(t, "emails", p.Base())
				assert.Equal(t, "type", p.FilterRoot().Left().Data().(Path).Base())
				assert.Equal(t, "value", p.Next().Base())
			},
		},
	} {
		p, err := NewPath(test.pathText)
		require.Nil(t, err)
		p.CorrectCase(sch, true)
		test.assertion(p)
	}
}
