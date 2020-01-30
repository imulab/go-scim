package expr

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestFilter(t *testing.T) {
	s := new(FilterTestSuite)
	suite.Run(t, s)
}

type FilterTestSuite struct {
	suite.Suite
}

func (s *FilterTestSuite) TestFilterCompiler() {
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
		filter string
		assert func(t *testing.T, trail []expect, err error)
	}{
		{
			name:   "simple filter",
			filter: "username eq \"foo\"",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 3)

				assert.Equal(t, Eq, trail[0].value)
				assert.Equal(t, "username", trail[1].value)
				assert.Equal(t, "\"foo\"", trail[2].value)

				assert.Equal(t, operator, trail[0].typ)
				assert.Equal(t, step, trail[1].typ)
				assert.Equal(t, literal, trail[2].typ)
			},
		},
		{
			name:   "simple filter with parenthesis",
			filter: "(age gt 10)",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 3)

				assert.Equal(t, Gt, trail[0].value)
				assert.Equal(t, "age", trail[1].value)
				assert.Equal(t, "10", trail[2].value)

				assert.Equal(t, operator, trail[0].typ)
				assert.Equal(t, step, trail[1].typ)
				assert.Equal(t, literal, trail[2].typ)
			},
		},
		{
			name:   "simple filter with urn prefix",
			filter: "urn:ietf:params:scim:schemas:core:2.0:User:meta.created gt \"2019-10-10T10:10:10\"",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 5)

				assert.Equal(t, Gt, trail[0].value)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User", trail[1].value)
				assert.Equal(t, "meta", trail[2].value)
				assert.Equal(t, "created", trail[3].value)
				assert.Equal(t, "\"2019-10-10T10:10:10\"", trail[4].value)

				assert.Equal(t, operator, trail[0].typ)
				assert.Equal(t, step, trail[1].typ)
				assert.Equal(t, step, trail[2].typ)
				assert.Equal(t, step, trail[3].typ)
				assert.Equal(t, literal, trail[4].typ)
			},
		},
		{
			name:   "filter starts with not operator",
			filter: "not (name pr)",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 3)

				assert.Equal(t, Not, trail[0].value)
				assert.Equal(t, Pr, trail[1].value)
				assert.Equal(t, "name", trail[2].value)

				assert.Equal(t, operator, trail[0].typ)
				assert.Equal(t, operator, trail[1].typ)
				assert.Equal(t, step, trail[2].typ)
			},
		},
		{
			name:   "composite filter",
			filter: "(username eq \"foo\") and (age gt 10)",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.Nil(t, err)
				assert.Len(t, trail, 7)

				assert.Equal(t, And, trail[0].value)
				assert.Equal(t, Eq, trail[1].value)
				assert.Equal(t, "username", trail[2].value)
				assert.Equal(t, "\"foo\"", trail[3].value)
				assert.Equal(t, Gt, trail[4].value)
				assert.Equal(t, "age", trail[5].value)
				assert.Equal(t, "10", trail[6].value)

				assert.Equal(t, operator, trail[0].typ)
				assert.Equal(t, operator, trail[1].typ)
				assert.Equal(t, step, trail[2].typ)
				assert.Equal(t, literal, trail[3].typ)
				assert.Equal(t, operator, trail[4].typ)
				assert.Equal(t, step, trail[5].typ)
				assert.Equal(t, literal, trail[6].typ)
			},
		},
		{
			name:   "invalid filter: starts with literal",
			filter: "\"hello\" eq false",
			assert: func(t *testing.T, trail []expect, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			root, err := CompileFilter(test.filter)
			if err != nil || root == nil {
				test.assert(t, nil, err)
			} else {
				trail := make([]expect, 0)
				root.Walk(func(step *Expression) {
					trail = append(trail, expect{
						value: step.token,
						typ:   selectType(step),
					})
				}, root, func() {
					test.assert(t, trail, err)
				})
			}
		})
	}
}

func (s *FilterTestSuite) TestFilterScanner() {
	type signals struct {
		event   int
		repeats int
	}

	tests := []struct {
		name   string
		filter string
		print  bool
		expect []signals
	}{
		{
			name:   "simple filter",
			filter: "username eq \"foo\"",
			expect: []signals{
				{scanFilterBeginPath, 1},
				{scanFilterContinue, 7},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
				{scanFilterBeginLiteral, 1},
				{scanFilterContinue, 4},
				{scanFilterEndLiteral, 1},
			},
		},
		{
			name:   "simple filter 2",
			filter: "(age gt 10)",
			expect: []signals{
				{scanFilterParenthesis, 1},
				{scanFilterBeginPath, 1},
				{scanFilterContinue, 2},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
				{scanFilterBeginLiteral, 1},
				{scanFilterContinue, 1},
				{scanFilterEndLiteral, 1},
				{scanFilterParenthesis, 1},
			},
		},
		{
			name:   "simple filter with parenthesis",
			filter: "(username eq \"foo\")",
			expect: []signals{
				{scanFilterParenthesis, 1},
				{scanFilterBeginPath, 1},
				{scanFilterContinue, 7},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
				{scanFilterBeginLiteral, 1},
				{scanFilterContinue, 4},
				{scanFilterEndLiteral, 1},
				{scanFilterParenthesis, 1},
			},
		},
		{
			name:   "simple filter with urn prefix",
			filter: "urn:ietf:params:scim:schemas:core:2.0:User:meta.created gt \"2019-10-10T10:10:10\"",
			expect: []signals{
				{scanFilterBeginPath, 1},
				{scanFilterContinue, 54},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
				{scanFilterBeginLiteral, 1},
				{scanFilterContinue, 20},
				{scanFilterEndLiteral, 1},
			},
		},
		{
			name:   "filter starts with not operator",
			filter: "not (name pr)",
			expect: []signals{
				{scanFilterBeginAny, 1},
				{scanFilterContinue, 2},
				{scanFilterEndOp, 1},
				{scanFilterParenthesis, 1},
				{scanFilterBeginAny, 1},
				{scanFilterContinue, 3},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
				{scanFilterParenthesis, 1},
				{scanFilterEnd, 1},
			},
		},
		{
			name:   "composite filter",
			filter: "(username eq \"foo\") and (age gt 10)",
			expect: []signals{
				{scanFilterParenthesis, 1},
				{scanFilterBeginPath, 1},
				{scanFilterContinue, 7},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
				{scanFilterBeginLiteral, 1},
				{scanFilterContinue, 4},
				{scanFilterEndLiteral, 1},
				{scanFilterParenthesis, 1},
				{scanFilterSkipSpace, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 2},
				{scanFilterEndOp, 1},
				{scanFilterParenthesis, 1},
				{scanFilterBeginPath, 1},
				{scanFilterContinue, 2},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
				{scanFilterBeginLiteral, 1},
				{scanFilterContinue, 1},
				{scanFilterEndLiteral, 1},
				{scanFilterParenthesis, 1},
				{scanFilterEnd, 1},
			},
		},
		{
			name:   "invalid filter: starts with literal",
			filter: "\"hello\" eq false",
			expect: []signals{
				{scanFilterError, 17},
			},
		},
		{
			name:   "invalid filter: bad literal",
			filter: "foo eq bar",
			expect: []signals{
				{scanFilterBeginPath, 1},
				{scanFilterContinue, 2},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
				{scanFilterError, 4},
			},
		},
		{
			name:   "single cardinality",
			filter: "id pr",
			expect: []signals{
				{scanFilterBeginPath, 1},
				{scanFilterContinue, 1},
				{scanFilterEndPath, 1},
				{scanFilterBeginOp, 1},
				{scanFilterContinue, 1},
				{scanFilterEndOp, 1},
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t0 *testing.T) {
			scan := &filterScanner{}
			scan.init()

			ops := make([]int, 0)

			for _, b := range append([]byte(test.filter), 0) {
				op := scan.step(scan, b)

				if op == scanFilterInsertSpace {
					op = scan.step(scan, ' ')
					ops = append(ops, op)
					if test.print {
						println(string(' '), explainFilterEvent(op))
					}

					op = scan.step(scan, b)
				}

				ops = append(ops, op)
				if test.print {
					println(string(b), explainFilterEvent(op))
				}
			}

			c := 0
			for _, k := range test.expect {
				for i := 0; i < k.repeats; i++ {
					assert.Equal(t0, k.event, ops[c])
					c++
				}
			}
		})
	}
}
