package mongo

import (
	"fmt"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/query"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"strconv"
	"time"
)

// Contrary to the main theme in this package, the methods in this file transforms SCIM filter to an
// intermediate format, which will again be transformed by the MongoDB driver to BSON. This decision
// was made based on two facts: first, filter is relatively small, hence the performance impact is negligible;
// second, operations involving filters are performed less frequent, hence, even if there is a performance
// penalty for double transformation and reflection, the overall throughput will not be affected as much.
//
// By transforming the query to bson.M (the map based intermediate format), we gain additional readability which
// is less achievable when directly appending BSON bytes to a buffer.

// Compile and transform a SCIM filter string to a bsonx.Val that contains the original
// filter in MongoDB compatible format.
func TransformFilter(scimFilter string, resourceType *core.ResourceType) (bsonx.Val, error) {
	root, err := query.CompileFilter(scimFilter)
	if err != nil {
		return bsonx.Null(), err
	}
	return TransformCompiledFilter(root, resourceType)
}

// Compile and transform a compiled SCIM filter to bsonx.Val that contains the original
// filter in MongoDB compatible format. This slight optimization allow the caller to pre-compile
// frequently used queries and save the trip to the filter parser and compiler.
func TransformCompiledFilter(root *core.Step, resourceType *core.ResourceType) (bsonx.Val, error) {
	return (&transformer{resourceType: resourceType}).transform(root)
}

type transformer struct {
	resourceType *core.ResourceType
}

// Transform the filter which is represented by the root to bsonx.Val.
func (t *transformer) transform(root *core.Step) (bsonx.Val, error) {
	var (
		left, right bsonx.Val
		err         error
	)

	// logical operators
	switch root.Token {
	case core.And:
		left, err = t.transform(root.Left)
		if err != nil {
			return bsonx.Null(), err
		}
		right, err = t.transform(root.Right)
		if err != nil {
			return bsonx.Null(), err
		}
		return bsonx.Document(bsonx.MDoc{
			mongoAnd: bsonx.Array(bsonx.Arr{left, right}),
		}), nil

	case core.Or:
		left, err = t.transform(root.Left)
		if err != nil {
			return bsonx.Null(), err
		}
		right, err = t.transform(root.Right)
		if err != nil {
			return bsonx.Null(), err
		}
		return bsonx.Document(bsonx.MDoc{
			mongoOr: bsonx.Array(bsonx.Arr{left, right}),
		}), nil

	case core.Not:
		left, err = t.transform(root.Left)
		if err != nil {
			return bsonx.Null(), err
		}
		return bsonx.Document(bsonx.MDoc{
			mongoNot: bsonx.Array(bsonx.Arr{left}),
		}), nil
	}

	var (
		path	string
		attr	*core.Attribute
		value	bsonx.Val
	)
	{
		path, attr, err = t.getFullMongoPathAndAttribute(root.Left)
		if err != nil {
			return bsonx.Null(), err
		}

		err = attr.CheckOpCompatibility(root.Token)
		if err != nil {
			return bsonx.Null(), core.Errors.InvalidFilter(err.Error())
		}

		value, err = t.parseValue(root.Right.Token, attr.ToSingleValued())
		if err != nil {
			return bsonx.Null(), err
		}
	}

	switch root.Token {
	case core.Eq:
		if attr.MultiValued {
			// Here, we are only allowing multiValued element match on eq operators, as the semantics of multiValued
			// match on other operators are not clear in the specification.
			if attr.Type != core.TypeString || attr.CaseExact {
				return bsonx.Document(bsonx.MDoc{
					path: bsonx.Document(bsonx.MDoc{
						mongoElementMatch: bsonx.Document(bsonx.MDoc{
							mongoEq: value,
						}),
					}),
				}), nil
			} else {
				return bsonx.Document(bsonx.MDoc{
					path: bsonx.Document(bsonx.MDoc{
						mongoElementMatch: bsonx.Array(bsonx.Arr{
							bsonx.Regex(fmt.Sprintf("^%s$", root.Right.Token), "i"),
						}),
					}),
				}), nil
			}
		} else {
			// caseExact option is only effective on string type:
			// - numeric and boolean type has no concept of case
			// - other string based types such as reference, binary, dateTime is not case sensitive
			if attr.Type != core.TypeString || attr.CaseExact {
				return bsonx.Document(bsonx.MDoc{
					path: bsonx.Document(bsonx.MDoc{
						mongoEq: value,
					}),
				}), nil
			} else {
				return bsonx.Document(bsonx.MDoc{
					path: bsonx.Regex(fmt.Sprintf("^%s$", root.Right.Token), "i"),
				}), nil
			}
		}

	case core.Ne:
		// caseExact option is only effective on string type:
		// - numeric and boolean type has no concept of case
		// - other string based types such as reference, binary, dateTime is not case sensitive
		if attr.Type != core.TypeString || attr.CaseExact {
			return bsonx.Document(bsonx.MDoc{
				path: bsonx.Document(bsonx.MDoc{
					mongoNe: value,
				}),
			}), nil
		} else {
			return bsonx.Document(bsonx.MDoc{
				mongoNot: bsonx.Array(bsonx.Arr{
					bsonx.Document(bsonx.MDoc{
						path: bsonx.Regex(fmt.Sprintf("^%s$", root.Right.Token), "i"),
					}),
				}),
			}), nil
		}

	case core.Sw:
		if attr.CaseExact {
			return bsonx.Document(bsonx.MDoc{
				path: bsonx.Regex(fmt.Sprintf("^%s", root.Right.Token), ""),
			}), nil
		} else {
			return bsonx.Document(bsonx.MDoc{
				path: bsonx.Regex(fmt.Sprintf("^%s", root.Right.Token), "i"),
			}), nil
		}

	case core.Ew:
		if attr.CaseExact {
			return bsonx.Document(bsonx.MDoc{
				path: bsonx.Regex(fmt.Sprintf("%s$", root.Right.Token), ""),
			}), nil
		} else {
			return bsonx.Document(bsonx.MDoc{
				path: bsonx.Regex(fmt.Sprintf("%s$", root.Right.Token), "i"),
			}), nil
		}

	case core.Co:
		if attr.CaseExact {
			return bsonx.Document(bsonx.MDoc{
				path: bsonx.Regex(root.Right.Token, ""),
			}), nil
		} else {
			return bsonx.Document(bsonx.MDoc{
				path: bsonx.Regex(root.Right.Token, "i"),
			}), nil
		}

	case core.Gt:
		return bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{mongoGt: value}),
		}), nil

	case core.Ge:
		return bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{mongoGe: value}),
		}), nil

	case core.Lt:
		return bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{mongoLt: value}),
		}), nil

	case core.Le:
		return bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{mongoLe: value}),
		}), nil

	case core.Pr:
		existsCriteria := bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{
				mongoExists: bsonx.Boolean(true),
			}),
		})
		nullCriteria := bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{
				mongoNe: bsonx.Null(),
			}),
		})
		emptyStringCriteria := bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{
				mongoNe: bsonx.String(""),
			}),
		})
		emptyArrayCriteria := bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{
				mongoNot: bsonx.Array(bsonx.Arr{
					bsonx.Document(bsonx.MDoc{mongoSize: bsonx.Int32(0)}),
				}),
			}),
		})
		emptyObjectCriteria := bsonx.Document(bsonx.MDoc{
			path: bsonx.Document(bsonx.MDoc{
				mongoNe: bsonx.Document(bsonx.MDoc{}),
			}),
		})

		criterion := make([]bsonx.Val, 0)
		criterion = append(criterion, existsCriteria, nullCriteria)
		if attr.MultiValued {
			criterion = append(criterion, emptyArrayCriteria)
		} else {
			switch attr.Type {
			case core.TypeString, core.TypeReference, core.TypeBinary:
				criterion = append(criterion, emptyStringCriteria)
			case core.TypeComplex:
				criterion = append(criterion, emptyObjectCriteria)
			}
		}
		return bsonx.Document(bsonx.MDoc{
			mongoAnd: bsonx.Array(bsonx.Arr(criterion)),
		}), nil
	}

	panic("invalid operator")
}

// Get a container attribute, containing all the derived attributes from the resource type.
func (t *transformer) getSuperAttribute() *core.Attribute {
	return &core.Attribute{
		SubAttributes: t.resourceType.DerivedAttributes(),
	}
}

// Parse the given raw value to the corresponding bsonx.Val according to the type information in attribute.
func (t transformer) parseValue(raw string, attr *core.Attribute) (bsonx.Val, error) {
	if attr.MultiValued || attr.Type == core.TypeComplex {
		panic("parse value only applicable to singleValued non-complex type")
	}

	switch attr.Type {
	case core.TypeString, core.TypeReference, core.TypeBinary:
		return bsonx.String(raw), nil
	case core.TypeDateTime:
		t, err := time.Parse(core.ISO8601, raw)
		if err != nil {
			return bsonx.Null(), core.Errors.InvalidFilter(fmt.Sprintf("invalid value: expects '%s' to be a dateTime", raw))
		}
		return bsonx.Time(t), nil
	case core.TypeBoolean:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return bsonx.Null(), core.Errors.InvalidFilter(fmt.Sprintf("invalid value: expects '%s' to be a boolean", raw))
		}
		return bsonx.Boolean(b), nil
	case core.TypeInteger:
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return bsonx.Null(), core.Errors.InvalidFilter(fmt.Sprintf("invalid value: expects '%s' to be an integer", raw))
		}
		return bsonx.Int64(i), nil
	case core.TypeDecimal:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return bsonx.Null(), core.Errors.InvalidFilter(fmt.Sprintf("invalid value: expects '%s' to be a decimal", raw))
		}
		return bsonx.Double(f), nil
	default:
		panic("impossible type")
	}
}

// Get the full mongo db path equivalent of the path represented by the step list, and the attribute of this path.
func (t transformer) getFullMongoPathAndAttribute(head *core.Step) (dbPath string, attr *core.Attribute, err error) {
	var cursor = head
	if cursor == nil {
		err = core.Errors.InvalidFilter("no such path in filter")
		return
	}

	dbPath = ""
	attr = t.getSuperAttribute()

path:
	for cursor != nil {
		if !cursor.IsPath() {
			err = core.Errors.InvalidFilter(fmt.Sprintf("'%s' is not a path", cursor.Token))
			return
		}

		for _, subAttr := range attr.SubAttributes {
			if subAttr.GoesBy(cursor.Token) {
				if len(dbPath) > 0 {
					dbPath += "."
				}
				// dbAlias take precedence over attribute name
				if subAttr.Metadata != nil && len(subAttr.Metadata.DbAlias) > 0 {
					dbPath += subAttr.Metadata.DbAlias
				} else {
					dbPath += subAttr.Name
				}

				// move to next step
				attr = subAttr
				cursor = cursor.Next
				continue path
			}
		}

		// exhaust all subAttributes => no such path
		err = core.Errors.InvalidFilter("no such path in filter")
		return
	}

	return
}

const (
	mongoAnd = "$and"
	mongoOr = "$or"
	mongoNot = "$nor"
	mongoElementMatch = "$elemMatch"
	mongoEq = "$eq"
	mongoNe = "$ne"
	mongoGt = "$gt"
	mongoGe = "$gte"
	mongoLt = "$lt"
	mongoLe = "$lte"
	mongoExists = "$exists"
	mongoSize = "$size"
)