package shared

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestComplex_Set(t *testing.T) {
	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.NotNil(t, sch)
	require.Nil(t, err)

	for _, test := range []struct {
		complex   Complex
		pathText  string
		value     interface{}
		assertion func(c Complex, err error)
	}{
		{
			// single path
			Complex{"userName": "foo"},
			"userName",
			"david",
			func(c Complex, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "david", c["userName"])
			},
		},
		{
			// duplex path
			Complex{"name": map[string]interface{}{"familyName": "Qiu"}},
			"name.familyName",
			"Q",
			func(c Complex, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "Q", c["name"].(map[string]interface{})["familyName"])
			},
		},
		{
			// duplex path (previously does not exist)
			Complex{"name": map[string]interface{}{"familyName": "Qiu"}},
			"name.givenName",
			"David",
			func(c Complex, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "David", c["name"].(map[string]interface{})["givenName"])
			},
		},
		{
			// simple path with filter
			Complex{"emails": []interface{}{
				map[string]interface{}{
					"value": "a@foo.com",
					"type":  "work",
				},
				map[string]interface{}{
					"value": "b@foo.com",
					"type":  "home",
				},
				map[string]interface{}{
					"value": "c@foo.com",
					"type":  "work",
				},
			}},
			"emails[type eq \"home\"]",
			map[string]interface{}{
				"value": "d@foo.com",
				"type":  "other",
			},
			func(c Complex, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 3, len(c["emails"].([]interface{})))
				assert.True(t, reflect.DeepEqual(c["emails"].([]interface{})[1], map[string]interface{}{
					"value": "d@foo.com",
					"type":  "other",
				}))
			},
		},
		{
			// duplex path with filter
			Complex{"emails": []interface{}{
				map[string]interface{}{
					"value": "a@foo.com",
					"type":  "work",
				},
				map[string]interface{}{
					"value": "b@foo.com",
					"type":  "home",
				},
				map[string]interface{}{
					"value": "c@foo.com",
					"type":  "work",
				},
			}},
			"emails[type eq \"home\"].type",
			"other",
			func(c Complex, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 3, len(c["emails"].([]interface{})))
				assert.Equal(t, "other", c["emails"].([]interface{})[1].(map[string]interface{})["type"])
			},
		},
	} {
		p, err := NewPath(test.pathText)
		require.Nil(t, err)

		err = test.complex.Set(p, test.value, sch)
		test.assertion(test.complex, err)
	}
}

func TestComplex_Get(t *testing.T) {
	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.NotNil(t, sch)
	require.Nil(t, err)

	for _, test := range []struct {
		complex   Complex
		pathText  string
		assertion func(result chan interface{})
	}{
		{
			// no existing path
			Complex{"userName": "david"},
			"dummy",
			func(result chan interface{}) {
				assert.Nil(t, <-result)
			},
		},
		{
			// single path
			Complex{"userName": "david"},
			"userName",
			func(result chan interface{}) {
				assert.Equal(t, "david", <-result)
			},
		},
		{
			// duplex path
			Complex{"name": map[string]interface{}{"familyName": "Qiu"}},
			"name.familyName",
			func(result chan interface{}) {
				assert.Equal(t, "Qiu", <-result)
			},
		},
		{
			// simple path with filter
			Complex{
				"emails": []interface{}{
					map[string]interface{}{
						"value": "A",
						"type":  "work",
					},
					map[string]interface{}{
						"value": "B",
						"type":  "home",
					},
					map[string]interface{}{
						"value": "C",
						"type":  "work",
					},
				},
			},
			"emails[type eq \"work\"]",
			func(result chan interface{}) {
				matches := make([]interface{}, 0)
				for elem := range result {
					matches = append(matches, elem)
				}
				assert.Equal(t, 2, len(matches))
				assert.Equal(t, "A", matches[0].(map[string]interface{})["value"])
				assert.Equal(t, "work", matches[0].(map[string]interface{})["type"])
				assert.Equal(t, "C", matches[1].(map[string]interface{})["value"])
				assert.Equal(t, "work", matches[1].(map[string]interface{})["type"])
			},
		},
		{
			// duplex path with filter
			Complex{
				"emails": []interface{}{
					map[string]interface{}{
						"value": "A",
						"type":  "work",
					},
					map[string]interface{}{
						"value": "B",
						"type":  "home",
					},
					map[string]interface{}{
						"value": "C",
						"type":  "work",
					},
				},
			},
			"emails[type eq \"work\"].value",
			func(result chan interface{}) {
				matches := make([]interface{}, 0)
				for elem := range result {
					matches = append(matches, elem)
				}
				assert.Equal(t, 2, len(matches))
				assert.Equal(t, "A", matches[0])
				assert.Equal(t, "C", matches[1])
			},
		},
		{
			// duplex path with filter (part missing)
			Complex{
				"emails": []interface{}{
					map[string]interface{}{
						"type": "work",
					},
					map[string]interface{}{
						"value": "B",
						"type":  "home",
					},
					map[string]interface{}{
						"value": "C",
						"type":  "work",
					},
				},
			},
			"emails[type eq \"work\"].value",
			func(result chan interface{}) {
				matches := make([]interface{}, 0)
				for elem := range result {
					matches = append(matches, elem)
				}
				assert.Equal(t, 1, len(matches))
				assert.Equal(t, "C", matches[0])
			},
		},
	} {
		p, err := NewPath(test.pathText)
		require.Nil(t, err)
		test.assertion(test.complex.Get(p, sch))
	}
}
