package expr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromValue(t *testing.T) {
	type expect struct {
		value string
		typ   exprType
	}
	tests := []struct {
		name   string
		value  interface{}
		assert func(t *testing.T, trail []expect, err error)
	}{
		{
			name: "simple-value",
			value: map[string]interface{}{
				"value": "x",
			},
			assert: func(t *testing.T, trail []expect, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, len(trail))
				assert.Equal(t, relationalOp, trail[0].typ)
				assert.Equal(t, Eq, trail[0].value)
				assert.Equal(t, path, trail[1].typ)
				assert.Equal(t, "value", trail[1].value)
				assert.Equal(t, literal, trail[2].typ)
				assert.Equal(t, "\"x\"", trail[2].value)
			},
		},
		{
			name: "complex-value",
			value: map[string]interface{}{
				"s": "x",
				"i": int64(45),
				"f": float64(2.3),
				"b": true,
			},
			assert: func(t *testing.T, trail []expect, err error) {
				require.NoError(t, err)
				require.Equal(t, 15, len(trail))
				for _, i := range []int{0, 4, 8} {
					assert.Equal(t, logicalOp, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Equal(t, And, trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{1, 5, 9, 12} {
					assert.Equal(t, relationalOp, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Equal(t, Eq, trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{2, 6, 10, 13} {
					assert.Equal(t, path, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Contains(t,
						[]string{"s", "i", "f", "b"}, trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{3, 7, 11, 14} {
					assert.Equal(t, literal, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Contains(t,
						[]string{"\"x\"", "\"45\"", "\"2.3\"", "\"true\""},
						trail[i].value, fmt.Sprintf("node %d", i))
				}
			},
		},
		{
			name: "nested-value",
			value: map[string]interface{}{
				"1": "x",
				"2": map[string]interface{}{
					"1": "x",
					"2": map[string]interface{}{
						"1": "x",
					},
				},
			},
			assert: func(t *testing.T, trail []expect, err error) {
				require.NoError(t, err)
				require.Equal(t, 11, len(trail))
				for _, i := range []int{0, 4} {
					assert.Equal(t, logicalOp, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Equal(t, And, trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{1, 5, 8} {
					assert.Equal(t, relationalOp, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Equal(t, Eq, trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{2, 6, 9} {
					assert.Equal(t, path, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Contains(t,
						[]string{"1", "2.1", "2.2.1"}, trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{3, 7, 10} {
					assert.Equal(t, literal, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Equal(t, "\"x\"", trail[i].value, fmt.Sprintf("node %d", i))
				}
			},
		},
		{
			name: "multiple values",
			value: []interface{}{
				map[string]interface{}{
					"value": int64(1),
				},
				map[string]interface{}{
					"value": int64(2),
				},
				map[string]interface{}{
					"value": int64(3),
				},
				map[string]interface{}{
					"value": int64(4),
				},
			},
			assert: func(t *testing.T, trail []expect, err error) {
				require.NoError(t, err)
				require.Equal(t, 15, len(trail))
				for _, i := range []int{0, 4, 8} {
					assert.Equal(t, logicalOp, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Equal(t, Or, trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{1, 5, 9, 12} {
					assert.Equal(t, relationalOp, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Equal(t, Eq, trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{2, 6, 10, 13} {
					assert.Equal(t, path, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Equal(t, "value", trail[i].value, fmt.Sprintf("node %d", i))
				}
				for _, i := range []int{3, 7, 11, 14} {
					assert.Equal(t, literal, trail[i].typ, fmt.Sprintf("node %d", i))
					assert.Contains(t, []string{"\"1\"", "\"2\"", "\"3\"", "\"4\""},
						trail[i].value, fmt.Sprintf("node %d", i))
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				head *Expression
				err  error
			)
			switch v := test.value.(type) {
			case map[string]interface{}:
				head, err = FromValue(v)
			case []interface{}:
				head, err = FromValueList(v)
			default:
				require.True(t, false, "invalid test case")
			}
			if err != nil || head == nil {
				test.assert(t, nil, err)
			} else {
				trail := make([]expect, 0)
				head.Walk(func(step *Expression) {
					trail = append(trail, expect{
						value: step.token,
						typ:   step.typ,
					})
				}, head, func() {
					test.assert(t, trail, err)
				})
			}
		})
	}
}
