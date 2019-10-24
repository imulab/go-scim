package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilterScanner(t *testing.T) {
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t0 *testing.T) {
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
