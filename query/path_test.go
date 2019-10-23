package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathScanner(t *testing.T) {
	RegisterPathNamespace("urn:ietf:params:scim:schemas:core:2.0:User")

	tests := []struct {
		name   string
		path   string
		print  bool
		expect [][]int
	}{
		{
			name: "simple path",
			path: "username",
			expect: [][]int{
				{scanPathBeginStep, 1},
				{scanPathContinue, 7},
				{scanPathEndStep, 1},
			},
		},
		{
			name: "duplex path",
			path: "meta.created",
			expect: [][]int{
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
			expect: [][]int{
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
			expect: [][]int{
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
			expect: [][]int{
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
			expect: [][]int{
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
		t.Run(test.name, func(t0 *testing.T) {
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
				expectOp := k[0]
				expectRepeat := k[1]
				for i := 0; i < expectRepeat; i++ {
					assert.Equal(t0, expectOp, ops[m])
					m++
				}
			}
		})
	}
}
