package mongo

import (
	. "github.com/davidiamyou/go-scim/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"testing"
)

func TestConvertToMongoQuery(t *testing.T) {
	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)
	require.NotNil(t, sch)
	sch.Attributes = append(sch.Attributes, &Attribute{
		Name:        "age",
		Type:        TypeInteger,
		MultiValued: false,
		Required:    false,
		CaseExact:   false,
		Mutability:  ReadWrite,
		Returned:    Default,
		Uniqueness:  None,
		Assist:      &Assist{JSONName: "age", Path: "age", FullPath: UserUrn + ":age"},
	})

	for _, test := range []struct {
		queryText string
		assertion func(result bson.M, err error)
	}{
		{
			"username eq \"david\" and age gt 17",
			func(result bson.M, err error) {
				assert.Nil(t, err)
				assert.True(t, reflect.DeepEqual(result, bson.M{
					"$and": []interface{}{
						bson.M{
							"userName": bson.M{
								"$regex": bson.RegEx{
									Pattern: "^david$",
									Options: "i",
								},
							},
						},
						bson.M{
							"age": bson.M{
								"$gt": int64(17),
							},
						},
					},
				}))
			},
		},
		{
			"addresses.locality sw \"Sh\"",
			func(result bson.M, err error) {
				assert.Nil(t, err)
				assert.True(t, reflect.DeepEqual(result, bson.M{
					"addresses.locality": bson.M{
						"$regex": bson.RegEx{
							Pattern: "^Sh",
							Options: "i",
						},
					},
				}))
			},
		},
		{
			"username sw 3",
			func(result bson.M, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			"active eq false",
			func(result bson.M, err error) {
				assert.Nil(t, err)
				assert.True(t, reflect.DeepEqual(result, bson.M{
					"active": bson.M{
						"$eq": false,
					},
				}))
			},
		},
	} {
		test.assertion(convertToMongoQuery(test.queryText, sch))
	}
}
