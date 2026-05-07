package expr

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestPath(t *testing.T) {
	s := new(PathTestSuite)
	suite.Run(t, s)
}

type PathTestSuite struct {
	suite.Suite
}

func (s *PathTestSuite) TestPathCompiler() {
	const (
		step = iota
		operator
		literal
		bad
	)
	type expect struct {
		value string
		typ   int
	}
	selectType := func(s *Expression) int {
		if s.IsPath() {
			return step
		} else if s.IsOperator() {
			return operator
		} else if s.IsLiteral() {
			return literal
		} else {
			return bad
		}
	}

	RegisterURN("urn:ietf:params:scim:schemas:core:2.0:User")

	tests := []struct {
		name   string
		path   string
		assert func(t *testing.T, trail []expect, err error)
	}{
		{
			name: "simple path",
			path: "username",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 1)
				assert.Equal(t, "username", trail[0].value)
				assert.Equal(t, step, trail[0].typ)
			},
		},
		{
			name: "duplex path",
			path: "meta.created",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 2)
				assert.Equal(t, "meta", trail[0].value)
				assert.Equal(t, "created", trail[1].value)
				assert.Equal(t, step, trail[0].typ)
				assert.Equal(t, step, trail[1].typ)
			},
		},
		{
			name: "path with urn namespace",
			path: "urn:ietf:params:scim:schemas:core:2.0:User:emails.primary",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 3)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User", trail[0].value)
				assert.Equal(t, "emails", trail[1].value)
				assert.Equal(t, "primary", trail[2].value)
				assert.Equal(t, step, trail[0].typ)
				assert.Equal(t, step, trail[1].typ)
				assert.Equal(t, step, trail[2].typ)
			},
		},
		{
			name: "simple path with filter",
			path: "emails[primary eq true]",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 4)
				assert.Equal(t, "emails", trail[0].value)
				assert.Equal(t, Eq, trail[1].value)
				assert.Equal(t, "primary", trail[2].value)
				assert.Equal(t, "true", trail[3].value)
				assert.Equal(t, step, trail[0].typ)
				assert.Equal(t, operator, trail[1].typ)
				assert.Equal(t, step, trail[2].typ)
				assert.Equal(t, literal, trail[3].typ)
			},
		},
		{
			name: "duplex path with filter",
			path: "emails[primary eq true].value",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 5)
				assert.Equal(t, "emails", trail[0].value)
				assert.Equal(t, Eq, trail[1].value)
				assert.Equal(t, "primary", trail[2].value)
				assert.Equal(t, "true", trail[3].value)
				assert.Equal(t, "value", trail[4].value)
				assert.Equal(t, step, trail[0].typ)
				assert.Equal(t, operator, trail[1].typ)
				assert.Equal(t, step, trail[2].typ)
				assert.Equal(t, literal, trail[3].typ)
				assert.Equal(t, step, trail[4].typ)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			head, err := CompilePath(test.path)
			if err != nil || head == nil {
				test.assert(t, nil, err)
			} else {
				trail := make([]expect, 0)
				head.Walk(func(step *Expression) {
					trail = append(trail, expect{
						value: step.token,
						typ:   selectType(step),
					})
				}, head, func() {
					test.assert(t, trail, err)
				})
			}
		})
	}
}

func (s *PathTestSuite) TestPathScanner() {
	type signals struct {
		event   int
		repeats int
	}

	RegisterURN("urn:ietf:params:scim:schemas:core:2.0:User")

	tests := []struct {
		name   string
		path   string
		print  bool
		expect []signals
	}{
		{
			name: "simple path",
			path: "username",
			expect: []signals{
				{scanPathBeginStep, 1},
				{scanPathContinue, 7},
				{scanPathEndStep, 1},
			},
		},
		{
			name: "duplex path",
			path: "meta.created",
			expect: []signals{
				{scanPathBeginStep, 1},
				{scanPathContinue, 3},
				{scanPathEndStep, 1},
				{scanPathBeginStep, 1},
				{scanPathContinue, 6},
				{scanPathEndStep, 1},
			},
		},
		{
			name: "simple path with urn prefix",
			path: "urn:ietf:params:scim:schemas:core:2.0:User:userName",
			expect: []signals{
				{scanPathBeginStep, 1},
				{scanPathContinue, 41},
				{scanPathEndStep, 1},
				{scanPathBeginStep, 1},
				{scanPathContinue, 7},
				{scanPathEndStep, 1},
			},
		},
		{
			name: "duplex path with urn prefix",
			path: "urn:ietf:params:scim:schemas:core:2.0:User:meta.created",
			expect: []signals{
				{scanPathBeginStep, 1},
				{scanPathContinue, 41},
				{scanPathEndStep, 1},
				{scanPathBeginStep, 1},
				{scanPathContinue, 3},
				{scanPathEndStep, 1},
				{scanPathBeginStep, 1},
				{scanPathContinue, 6},
				{scanPathEndStep, 1},
			},
		},
		{
			name: "simple path with filter",
			path: "emails[primary eq true]",
			expect: []signals{
				{scanPathBeginStep, 1},
				{scanPathContinue, 5},
				{scanPathBeginFilter, 1},
				{scanPathContinue, 15},
				{scanPathEndFilter, 1},
				{scanPathEnd, 1},
			},
		},
		{
			name: "duplex path with filter",
			path: "emails[primary eq true].value",
			expect: []signals{
				{scanPathBeginStep, 1},
				{scanPathContinue, 5},
				{scanPathBeginFilter, 1},
				{scanPathContinue, 15},
				{scanPathEndFilter, 1},
				{scanPathContinue, 1},
				{scanPathBeginStep, 1},
				{scanPathContinue, 4},
				{scanPathEndStep, 1},
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t0 *testing.T) {
			ops := make([]int, 0)
			scan := &pathScanner{}
			scan.init()

			for _, c := range append([]byte(test.path), 0) {
				op := scan.step(scan, c)
				if test.print {
					println(string(c), explainPathEvent(op))
				}
				ops = append(ops, op)
			}

			m := 0
			for _, k := range test.expect {
				for i := 0; i < k.repeats; i++ {
					assert.Equal(t0, k.event, ops[m])
					m++
				}
			}
		})
	}
}
