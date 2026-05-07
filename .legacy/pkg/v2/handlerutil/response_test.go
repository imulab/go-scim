package handlerutil

import (
	"errors"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect func(t *testing.T, raw []byte)
	}{
		{
			name: "wrapped scim error",
			err:  fmt.Errorf("%w: valid is invalid", spec.ErrInvalidValue),
			expect: func(t *testing.T, raw []byte) {
				assert.JSONEq(t, `
{
  "schemas": [
    "urn:ietf:params:scim:api:messages:2.0:Error"
  ],
  "status": 400,
  "scimType": "invalidValue",
  "detail": "invalidValue: valid is invalid"
}
`, string(raw))
			},
		},
		{
			name: "non scim error",
			err:  errors.New("something was wrong"),
			expect: func(t *testing.T, raw []byte) {
				assert.JSONEq(t, `
{
  "schemas":[
    "urn:ietf:params:scim:api:messages:2.0:Error"
  ],
  "status":500,
  "scimType":"internal",
  "detail":"something was wrong"
}
`, string(raw))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			assert.Nil(t, WriteError(rw, test.err))
			test.expect(t, rw.Body.Bytes())
		})
	}
}
