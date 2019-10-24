package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathScanner(t *testing.T) {
	type signals struct {
		event   int
		repeats int
	}

	RegisterPathNamespace("urn:ietf:params:scim:schemas:core:2.0:User")

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
				for i := 0; i < k.repeats; i++ {
					assert.Equal(t0, k.event, ops[m])
					m++
				}
			}
		})
	}
}
