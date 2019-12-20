package mongo

import (
	"fmt"
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/core/expr"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"strconv"
	"strings"
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
func TransformFilter(scimFilter string, resourceType *spec.ResourceType) (bsonx.Val, error) {
	root, err := expr.CompileFilter(scimFilter)
	if err != nil {
		return bsonx.Null(), err
	}
	return TransformCompiledFilter(root, resourceType)
}

// Compile and transform a compiled SCIM filter to bsonx.Val that contains the original
// filter in MongoDB compatible format. This slight optimization allow the caller to pre-compile
// frequently used queries and save the trip to the filter parser and compiler.
func TransformCompiledFilter(root *expr.Expression, resourceType *spec.ResourceType) (bsonx.Val, error) {
	return newTransformer(resourceType).transform(root)
}

func newTransformer(resourceType *spec.ResourceType) *transformer {
	return &transformer{
		superAttr: resourceType.SuperAttribute(true),
	}
}

type transformer struct {
	superAttr *spec.Attribute
}

// Transform the filter which is represented by the root to bsonx.Val.
func (t *transformer) transform(root *expr.Expression) (bsonx.Val, error) {
	switch root.Token() {
	case expr.And:
		return t.transformAnd(root)
	case expr.Or:
		return t.transformOr(root)
	case expr.Not:
		return t.transformNot(root)
	default:
		return t.transformRelational(t.superAttr, root.Left(), root, root.Right())
	}
}

func (t *transformer) transformAnd(root *expr.Expression) (bsonx.Val, error) {
	left, err := t.transform(root.Left())
	if err != nil {
		return bsonx.Null(), err
	}
	right, err := t.transform(root.Right())
	if err != nil {
		return bsonx.Null(), err
	}
	return bsonx.Document(bsonx.MDoc{
		mongoAnd: bsonx.Array(bsonx.Arr{left, right}),
	}), nil
}

func (t *transformer) transformOr(root *expr.Expression) (bsonx.Val, error) {
	left, err := t.transform(root.Left())
	if err != nil {
		return bsonx.Null(), err
	}
	right, err := t.transform(root.Right())
	if err != nil {
		return bsonx.Null(), err
	}
	return bsonx.Document(bsonx.MDoc{
		mongoOr: bsonx.Array(bsonx.Arr{left, right}),
	}), nil
}

func (t *transformer) transformNot(root *expr.Expression) (bsonx.Val, error) {
	left, err := t.transform(root.Left())
	if err != nil {
		return bsonx.Null(), err
	}
	return bsonx.Document(bsonx.MDoc{
		mongoNot: bsonx.Array(bsonx.Arr{left}),
	}), nil
}

func (t *transformer) transformRelational(containerAttr *spec.Attribute, path *expr.Expression, op *expr.Expression, value *expr.Expression) (bsonx.Val, error) {
	var (
		cursorAttr = containerAttr
		pathNames  = make([]string, 0)
	)
	{
		for path != nil && cursorAttr.SingleValued() {
			cursorAttr = cursorAttr.SubAttributeForName(path.Token())
			if cursorAttr == nil {
				return bsonx.Null(), errors.InvalidFilter("no such path in filter")
			}

			pathName := cursorAttr.Name()
			if md, ok := metadataHub[cursorAttr.ID()]; ok {
				pathName = md.MongoName
			}
			pathNames = append(pathNames, pathName)

			path = path.Next()
		}
	}

	var nextDoc bsonx.Val
	{
		var err error
		if path == nil {
			nextDoc, err = t.transformValue(cursorAttr, op, value)
		} else {
			nextDoc, err = t.transformRelational(cursorAttr.NewElementAttribute(), path, op, value)
		}
		if err != nil {
			return bsonx.Null(), err
		}
	}

	if cursorAttr.MultiValued() && (path != nil || op.Token() != expr.Pr) {
		// If we have stopped on a multiValued attribute, we do an $elementMatch when
		// 1. there's more to the path
		// 		i.e. emails.value
		// 2. the operator is not a pr
		// 		i.e. schemas eq "foobar"
		return bsonx.Document(bsonx.MDoc{
			strings.Join(pathNames, "."): bsonx.Document(bsonx.MDoc{
				mongoElementMatch: nextDoc,
			}),
		}), nil
	} else {
		return bsonx.Document(bsonx.MDoc{
			strings.Join(pathNames, "."): nextDoc,
		}), nil
	}
}

func (t *transformer) eqValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	if attr.Type() != spec.TypeString && attr.CaseExact() {
		return bsonx.Document(bsonx.MDoc{mongoEq: bsonx.String(unquote(value.Token()))}), nil
	} else {
		return bsonx.Regex(fmt.Sprintf("^%s$", unquote(value.Token())), "i"), nil
	}
}

func (t *transformer) neValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	if attr.Type() != spec.TypeString || attr.CaseExact() {
		return bsonx.Document(bsonx.MDoc{mongoNe: bsonx.String(unquote(value.Token()))}), nil
	} else {
		return bsonx.Regex(fmt.Sprintf("^((?!%s$).)", unquote(value.Token())), "i"), nil
	}
}

func (t *transformer) swValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	if attr.CaseExact() {
		return bsonx.Regex(fmt.Sprintf("^%s", unquote(value.Token())), ""), nil
	} else {
		return bsonx.Regex(fmt.Sprintf("^%s", unquote(value.Token())), "i"), nil
	}
}

func (t *transformer) ewValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	if attr.CaseExact() {
		return bsonx.Regex(fmt.Sprintf("%s$", unquote(value.Token())), ""), nil
	} else {
		return bsonx.Regex(fmt.Sprintf("%s$", unquote(value.Token())), "i"), nil
	}
}

func (t *transformer) coValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	if attr.CaseExact() {
		return bsonx.Regex(unquote(value.Token()), ""), nil
	} else {
		return bsonx.Regex(unquote(value.Token()), "i"), nil
	}
}

func (t *transformer) gtValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	v, err := t.parseValue(value.Token(), attr)
	if err != nil {
		return bsonx.Null(), err
	}
	return bsonx.Document(bsonx.MDoc{mongoGt: v}), nil
}

func (t *transformer) geValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	v, err := t.parseValue(value.Token(), attr)
	if err != nil {
		return bsonx.Null(), err
	}
	return bsonx.Document(bsonx.MDoc{mongoGe: v}), nil
}

func (t *transformer) ltValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	v, err := t.parseValue(value.Token(), attr)
	if err != nil {
		return bsonx.Null(), err
	}
	return bsonx.Document(bsonx.MDoc{mongoLt: v}), nil
}

func (t *transformer) leValue(attr *spec.Attribute, value *expr.Expression) (bsonx.Val, error) {
	v, err := t.parseValue(value.Token(), attr)
	if err != nil {
		return bsonx.Null(), err
	}
	return bsonx.Document(bsonx.MDoc{mongoLe: v}), nil
}

func (t *transformer) prDoc(attr *spec.Attribute) (bsonx.Val, error) {
	criterion := make([]bsonx.Val, 0)
	criterion = append(criterion, existsCriteria, nullCriteria)
	if attr.MultiValued() {
		criterion = append(criterion, emptyArrayCriteria)
	} else {
		switch attr.Type() {
		case spec.TypeString, spec.TypeReference, spec.TypeBinary:
			criterion = append(criterion, emptyStringCriteria)
		case spec.TypeComplex:
			criterion = append(criterion, emptyObjectCriteria)
		}
	}
	return bsonx.Document(bsonx.MDoc{
		mongoAnd: bsonx.Array(criterion),
	}), nil
}

func (t *transformer) transformValue(attr *spec.Attribute, op *expr.Expression, value *expr.Expression) (bsonx.Val, error) {
	switch op.Token() {
	case expr.Eq:
		return t.eqValue(attr, value)
	case expr.Ne:
		return t.neValue(attr, value)
	case expr.Sw:
		return t.swValue(attr, value)
	case expr.Ew:
		return t.ewValue(attr, value)
	case expr.Co:
		return t.coValue(attr, value)
	case expr.Gt:
		return t.gtValue(attr, value)
	case expr.Ge:
		return t.geValue(attr, value)
	case expr.Lt:
		return t.ltValue(attr, value)
	case expr.Le:
		return t.leValue(attr, value)
	case expr.Pr:
		return t.prDoc(attr)
	default:
		panic("invalid relational operator")
	}
}

// Parse the given raw value to the corresponding bsonx.Val according to the type information in attribute.
// The attribute will be treated as singleValued even if it is multiValued.
func (t transformer) parseValue(raw string, attr *spec.Attribute) (bsonx.Val, error) {
	if attr.Type() == spec.TypeComplex {
		return bsonx.Null(), errors.InvalidFilter("incompatible complex property")
	}

	switch attr.Type() {
	case spec.TypeString, spec.TypeReference, spec.TypeBinary:

		return bsonx.String(unquote(raw)), nil
	case spec.TypeDateTime:
		t, err := time.Parse(prop.ISO8601, unquote(raw))
		if err != nil {
			return bsonx.Null(), errors.InvalidFilter("invalid value: expects '%s' to be a dateTime", raw)
		}
		return bsonx.Time(t), nil
	case spec.TypeBoolean:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return bsonx.Null(), errors.InvalidFilter("invalid value: expects '%s' to be a boolean", raw)
		}
		return bsonx.Boolean(b), nil
	case spec.TypeInteger:
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return bsonx.Null(), errors.InvalidFilter("invalid value: expects '%s' to be an integer", raw)
		}
		return bsonx.Int64(i), nil
	case spec.TypeDecimal:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return bsonx.Null(), errors.InvalidFilter("invalid value: expects '%s' to be a decimal", raw)
		}
		return bsonx.Double(f), nil
	default:
		panic("impossible type")
	}
}

func unquote(raw string) string {
	raw = strings.TrimPrefix(raw, "\"")
	raw = strings.TrimSuffix(raw, "\"")
	return raw
}

// pre compiled criteria
var (
	existsCriteria      = bsonx.Document(bsonx.MDoc{mongoExists: bsonx.Boolean(true)})
	nullCriteria        = bsonx.Document(bsonx.MDoc{mongoNe: bsonx.Null()})
	emptyStringCriteria = bsonx.Document(bsonx.MDoc{mongoNe: bsonx.String("")})
	emptyObjectCriteria = bsonx.Document(bsonx.MDoc{mongoNe: bsonx.Document(bsonx.MDoc{})})
	emptyArrayCriteria  = bsonx.Document(bsonx.MDoc{
		mongoNot: bsonx.Array(bsonx.Arr{
			bsonx.Document(bsonx.MDoc{mongoSize: bsonx.Int32(0)}),
		}),
	})
)

const (
	mongoAnd          = "$and"
	mongoOr           = "$or"
	mongoNot          = "$nor"
	mongoElementMatch = "$elemMatch"
	mongoEq           = "$eq"
	mongoNe           = "$ne"
	mongoGt           = "$gt"
	mongoGe           = "$gte"
	mongoLt           = "$lt"
	mongoLe           = "$lte"
	mongoExists       = "$exists"
	mongoSize         = "$size"
)
