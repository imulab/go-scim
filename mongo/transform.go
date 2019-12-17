package mongo

import (
	"fmt"
	. "github.com/parsable/go-scim/shared"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"sync"
)

func convertToMongoQuery(query string, guide AttributeSource) (m bson.M, err error) {
	m = bson.M{}

	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			default:
				err = Error.Text("%v", r)
			}
		}
	}()

	q, err := NewFilter(query)
	if err != nil {
		return
	}
	q.CorrectCase(guide)

	m = transformInstance.do(q, guide)
	err = nil
	return
}

var (
	singleTransform   sync.Once
	transformInstance *transform
)

func init() {
	singleTransform.Do(func() {
		transformInstance = &transform{}
	})
}

type transform struct{}

func (t *transform) do(root FilterNode, guide AttributeSource) bson.M {
	var (
		left, right bson.M
		attr        *Attribute
	)

	switch root.Type() {
	case LogicalOperator:
		if root.Left() != nil {
			left = t.do(root.Left(), guide)
		}
		if root.Right() != nil {
			right = t.do(root.Right(), guide)
		}

	case RelationalOperator:
		path := root.Left().Data().(Path)
		attr = guide.GetAttribute(path, true)
		if attr == nil {
			t.throwIfError(Error.NoAttribute(path.CollectValue()))
		}

		if attr.ExpectsComplex() && root.Data() != Pr {
			t.throwIfError(Error.InvalidFilter("", fmt.Sprintf("Cannot perform %v on complex attribute", root.Data())))
		}

		switch root.Data() {
		case Ge, Gt, Le, Lt:
			if attr.ExpectsBool() || attr.ExpectsBinary() {
				t.throwIfError(Error.InvalidFilter("", fmt.Sprintf("Cannot determine order on %s attribute", attr.Type)))
			}
		}
	}

	switch root.Data() {
	case And:
		return bson.M{
			"$and": []interface{}{left, right},
		}
	case Or:
		return bson.M{
			"$or": []interface{}{left, right},
		}
	case Not:
		return bson.M{
			"$nor": []interface{}{left},
		}
	case Eq:
		if !attr.ExpectsString() || attr.CaseExact {
			return bson.M{
				attr.Assist.Path: bson.M{
					"$eq": root.Right().Data(),
				},
			}
		} else {
			return bson.M{
				attr.Assist.Path: bson.M{
					"$regex": bson.RegEx{
						Pattern: fmt.Sprintf("^%v$", root.Right().Data()),
						Options: "i",
					},
				},
			}
		}
	case Ne:
		if !attr.ExpectsString() || attr.CaseExact {
			return bson.M{
				attr.Assist.Path: bson.M{
					"ne": root.Right().Data(),
				},
			}
		} else {
			return bson.M{
				"$nor": []interface{}{
					bson.M{
						attr.Assist.Path: bson.M{
							"$regex": bson.RegEx{
								Pattern: fmt.Sprintf("^%v$", root.Right().Data()),
								Options: "i",
							},
						},
					},
				},
			}
		}
	case Co:
		if !attr.ExpectsString() {
			t.throwIfError(Error.InvalidFilter("", "Cannot use co operator on non-string attributes."))
		} else if reflect.ValueOf(root.Right().Data()).Kind() != reflect.String {
			t.throwIfError(Error.InvalidFilter("", "Cannot use co operator with non-string value."))
		} else {
			if attr.CaseExact {
				return bson.M{
					attr.Assist.Path: bson.M{
						"$regex": bson.RegEx{
							Pattern: root.Right().Data().(string),
						},
					},
				}
			} else {
				return bson.M{
					attr.Assist.Path: bson.M{
						"$regex": bson.RegEx{
							Pattern: root.Right().Data().(string),
							Options: "i",
						},
					},
				}
			}
		}
	case Sw:
		if !attr.ExpectsString() {
			t.throwIfError(Error.InvalidFilter("", "Cannot use sw operator on non-string attributes."))
		} else if reflect.ValueOf(root.Right().Data()).Kind() != reflect.String {
			t.throwIfError(Error.InvalidFilter("", "Cannot use sw operator with non-string value."))
		} else {
			if attr.CaseExact {
				return bson.M{
					attr.Assist.Path: bson.M{
						"$regex": bson.RegEx{
							Pattern: "^" + root.Right().Data().(string),
						},
					},
				}
			} else {
				return bson.M{
					attr.Assist.Path: bson.M{
						"$regex": bson.RegEx{
							Pattern: "^" + root.Right().Data().(string),
							Options: "i",
						},
					},
				}
			}
		}
	case Ew:
		if !attr.ExpectsString() {
			t.throwIfError(Error.InvalidFilter("", "Cannot use ew operator on non-string attributes."))
		} else if reflect.ValueOf(root.Right().Data()).Kind() != reflect.String {
			t.throwIfError(Error.InvalidFilter("", "Cannot use ew operator with non-string value."))
		} else {
			if attr.CaseExact {
				return bson.M{
					attr.Assist.Path: bson.M{
						"$regex": bson.RegEx{
							Pattern: root.Right().Data().(string) + "$",
						},
					},
				}
			} else {
				return bson.M{
					attr.Assist.Path: bson.M{
						"$regex": bson.RegEx{
							Pattern: root.Right().Data().(string) + "$",
							Options: "i",
						},
					},
				}
			}
		}
	case Gt:
		return bson.M{
			attr.Assist.Path: bson.M{
				"$gt": root.Right().Data(),
			},
		}
	case Ge:
		return bson.M{
			attr.Assist.Path: bson.M{
				"$gte": root.Right().Data(),
			},
		}
	case Lt:
		return bson.M{
			attr.Assist.Path: bson.M{
				"$lt": root.Right().Data(),
			},
		}
	case Le:
		return bson.M{
			attr.Assist.Path: bson.M{
				"$lte": root.Right().Data(),
			},
		}
	case Pr:
		existsCriteria := bson.M{attr.Assist.Path: bson.M{"$exists": true}}
		nullCriteria := bson.M{attr.Assist.Path: bson.M{"$ne": nil}}
		emptyStringCriteria := bson.M{attr.Assist.Path: bson.M{"$ne": ""}}
		emptyArrayCriteria := bson.M{attr.Assist.Path: bson.M{"$not": bson.M{"$size": 0}}}
		emptyObjectCriteria := bson.M{attr.Assist.Path: bson.M{"$ne": bson.M{}}}

		criterion := make([]interface{}, 0)
		criterion = append(criterion, existsCriteria, nullCriteria)
		if attr.MultiValued {
			criterion = append(criterion, emptyArrayCriteria)
		} else {
			switch attr.Type {
			case TypeString:
				criterion = append(criterion, emptyStringCriteria)
			case TypeComplex:
				criterion = append(criterion, emptyObjectCriteria)
			}
		}
		return bson.M{"$and": criterion}
	default:
		t.throwIfError(Error.InvalidFilter("", fmt.Sprintf("unknown operator %v", root.Data())))
	}

	return nil
}

func (t *transform) throwIfError(err error) {
	if err != nil {
		panic(err)
	}
}
