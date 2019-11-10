package json

import "testing"

func TestScanner(t *testing.T) {
	scan := &scanner{}
	scan.reset()

	json := `{
	"schemas": ["urn1", "urn2"],
	"id": "9F87882C-BE11-46B7-952B-2518E5296F55",
	"meta": {
		"age": 18,
		"score": 123.5,
		"active": true
	},
	"emails": [
		{
			"value": "foo",
			"primary": true
		},
		{
			"value": "bar"
		}
	]
}`

	explainOp := func(op int) string {
		switch op {
		case scanContinue:
			return "continue"
		case scanBeginLiteral:
			return "begin literal"
		case scanBeginObject:
			return "begin object"
		case scanObjectKey:
			return "object key"
		case scanObjectValue:
			return "object value"
		case scanEndObject:
			return "end object"
		case scanBeginArray:
			return "begin array"
		case scanArrayValue:
			return "array value"
		case scanEndArray:
			return "end array"
		case scanSkipSpace:
			return "skip space"
		case scanEnd:
			return "end"
		case scanError:
			return "errInvalidSyntax"
		default:
			return "unknown"
		}
	}

	for i, c := range []byte(json) {
		op := scan.step(scan, c)
		println(i, string(c), explainOp(op))
	}
}
