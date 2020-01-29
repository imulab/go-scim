package v2

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func TransformFilter(scimFilter string, resourceType *spec.ResourceType) (bson.D, error) {
	root, err := expr.CompileFilter(scimFilter)
	if err != nil {
		return nil, err
	}
	return TransformCompiledFilter(root, resourceType)
}

// Compile and transform a compiled SCIM filter to bsonx.Val that contains the original
// filter in MongoDB compatible format. This slight optimization allow the caller to pre-compile
// frequently used queries and save the trip to the filter parser and compiler.
func TransformCompiledFilter(root *expr.Expression, resourceType *spec.ResourceType) (bson.D, error) {
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
func (t *transformer) transform(root *expr.Expression) (bson.D, error) {
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

func (t *transformer) transformAnd(root *expr.Expression) (bson.D, error) {
	left, err := t.transform(root.Left())
	if err != nil {
		return nil, err
	}
	right, err := t.transform(root.Right())
	if err != nil {
		return nil, err
	}
	return bson.D{
		{Key: mongoAnd, Value: bson.A{left, right}},
	}, nil
}

func (t *transformer) transformOr(root *expr.Expression) (bson.D, error) {
	left, err := t.transform(root.Left())
	if err != nil {
		return nil, err
	}
	right, err := t.transform(root.Right())
	if err != nil {
		return nil, err
	}
	return bson.D{
		{Key: mongoOr, Value: bson.A{left, right}},
	}, nil
}

func (t *transformer) transformNot(root *expr.Expression) (bson.D, error) {
	left, err := t.transform(root.Left())
	if err != nil {
		return nil, err
	}
	return bson.D{
		{Key: mongoNot, Value: bson.A{left}},
	}, nil
}

func (t *transformer) transformRelational(containerAttr *spec.Attribute, path *expr.Expression, op *expr.Expression, value *expr.Expression) (bson.D, error) {
	var (
		cursorAttr = containerAttr
		pathNames  = make([]string, 0)
	)
	{
		for path != nil && !cursorAttr.MultiValued() {
			cursorAttr = cursorAttr.SubAttributeForName(path.Token())
			if cursorAttr == nil {
				return nil, fmt.Errorf("%w: no path for '%s'", spec.ErrInvalidFilter, path.Token())
			}

			pathName := cursorAttr.Name()
			if md, ok := metadataHub[cursorAttr.ID()]; ok {
				pathName = md.MongoName
			}
			pathNames = append(pathNames, pathName)

			path = path.Next()
		}
	}

	var nextDoc interface{}
	{
		var err error
		if path == nil {
			nextDoc, err = t.transformValue(cursorAttr, op, value)
		} else {
			nextDoc, err = t.transformRelational(cursorAttr.DeriveElementAttribute(), path, op, value)
		}
		if err != nil {
			return nil, err
		}
	}

	if cursorAttr.MultiValued() && (path != nil || op.Token() != expr.Pr) {
		// If we have stopped on a multiValued attribute, we do an $elementMatch when
		// 1. there's more to the path
		// 		i.e. emails.value
		// 2. the operator is not a pr
		// 		i.e. schemas eq "foobar"
		return bson.D{
			{Key: strings.Join(pathNames, "."), Value: bson.D{
				{Key: mongoElementMatch, Value: nextDoc},
			}},
		}, nil
	} else {
		q := bson.D{{Key: strings.Join(pathNames, "."), Value: nextDoc}}
		if op.Token() != expr.Pr {
			return q, nil
		}
		return t.rearrangeForPr(q), nil
	}
}

// rearrange query in the form of "{ <field> : { $and : [ <criteria1>, <criteria2> , ... , <criteriaN>] }}", into
// { $and : [ { <field> : <criteria1> }, { <field> : <criteria2> }, ..., { <field> : <criteriaN> } ] }
func (t *transformer) rearrangeForPr(doc bson.D) bson.D {
	if len(doc) != 1 {
		return doc
	}

	field := doc[0].Key
	if len(field) == 0 {
		return doc
	}

	if _, ok := doc[0].Value.(bson.D); !ok {
		return doc
	} else if len(doc[0].Value.(bson.D)) != 1 {
		return doc
	} else if doc[0].Value.(bson.D)[0].Key != mongoAnd {
		return doc
	} else if _, ok := doc[0].Value.(bson.D)[0].Value.(bson.A); !ok {
		return doc
	}

	criterion := doc[0].Value.(bson.D)[0].Value.(bson.A)

	newCriterion := bson.A{}
	for _, c := range criterion {
		newCriterion = append(newCriterion, bson.D{{Key: field, Value: c}})
	}

	return bson.D{{Key: mongoAnd, Value: newCriterion}}
}

func (t *transformer) eqValue(attr *spec.Attribute, value *expr.Expression) interface{} {
	if attr.Type() != spec.TypeString && attr.CaseExact() {
		return bson.D{
			{Key: mongoEq, Value: unquote(value.Token())},
		}
	} else {
		return primitive.Regex{
			Pattern: fmt.Sprintf("^%s$", unquote(value.Token())),
			Options: "i",
		}
	}
}

func (t *transformer) neValue(attr *spec.Attribute, value *expr.Expression) interface{} {
	if attr.Type() != spec.TypeString || attr.CaseExact() {
		return bson.D{
			{Key: mongoNe, Value: unquote(value.Token())},
		}
	} else {
		return primitive.Regex{
			Pattern: fmt.Sprintf("^((?!%s$).)", unquote(value.Token())),
			Options: "i",
		}
	}
}

func (t *transformer) swValue(attr *spec.Attribute, value *expr.Expression) primitive.Regex {
	if attr.CaseExact() {
		return primitive.Regex{
			Pattern: fmt.Sprintf("^%s", unquote(value.Token())),
		}
	} else {
		return primitive.Regex{
			Pattern: fmt.Sprintf("^%s", unquote(value.Token())),
			Options: "i",
		}
	}
}

func (t *transformer) ewValue(attr *spec.Attribute, value *expr.Expression) primitive.Regex {
	if attr.CaseExact() {
		return primitive.Regex{
			Pattern: fmt.Sprintf("%s$", unquote(value.Token())),
		}
	} else {
		return primitive.Regex{
			Pattern: fmt.Sprintf("%s$", unquote(value.Token())),
			Options: "i",
		}
	}
}

func (t *transformer) coValue(attr *spec.Attribute, value *expr.Expression) primitive.Regex {
	if attr.CaseExact() {
		return primitive.Regex{
			Pattern: unquote(value.Token()),
		}
	} else {
		return primitive.Regex{
			Pattern: "unquote(value.Token())",
			Options: "i",
		}
	}
}

func (t *transformer) gtValue(attr *spec.Attribute, value *expr.Expression) (bson.D, error) {
	v, err := t.parseValue(value.Token(), attr)
	if err != nil {
		return nil, err
	}
	return bson.D{
		{Key: mongoGt, Value: v},
	}, nil
}

func (t *transformer) geValue(attr *spec.Attribute, value *expr.Expression) (bson.D, error) {
	v, err := t.parseValue(value.Token(), attr)
	if err != nil {
		return nil, err
	}
	return bson.D{
		{Key: mongoGe, Value: v},
	}, nil
}

func (t *transformer) ltValue(attr *spec.Attribute, value *expr.Expression) (bson.D, error) {
	v, err := t.parseValue(value.Token(), attr)
	if err != nil {
		return nil, err
	}
	return bson.D{
		{Key: mongoLt, Value: v},
	}, nil
}

func (t *transformer) leValue(attr *spec.Attribute, value *expr.Expression) (bson.D, error) {
	v, err := t.parseValue(value.Token(), attr)
	if err != nil {
		return nil, err
	}
	return bson.D{
		{Key: mongoLe, Value: v},
	}, nil
}

func (t *transformer) prDoc(attr *spec.Attribute) bson.D {
	criterion := bson.A{}
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
	return bson.D{{Key: mongoAnd, Value: criterion}}
}

func (t *transformer) transformValue(attr *spec.Attribute, op *expr.Expression, value *expr.Expression) (interface{}, error) {
	switch op.Token() {
	case expr.Eq:
		return t.eqValue(attr, value), nil
	case expr.Ne:
		return t.neValue(attr, value), nil
	case expr.Sw:
		return t.swValue(attr, value), nil
	case expr.Ew:
		return t.ewValue(attr, value), nil
	case expr.Co:
		return t.coValue(attr, value), nil
	case expr.Gt:
		return t.gtValue(attr, value)
	case expr.Ge:
		return t.geValue(attr, value)
	case expr.Lt:
		return t.ltValue(attr, value)
	case expr.Le:
		return t.leValue(attr, value)
	case expr.Pr:
		return t.prDoc(attr), nil
	default:
		panic("invalid relational operator")
	}
}

func (t transformer) errIncompatibleValue(attr *spec.Attribute) error {
	return fmt.Errorf("%w: value in filter incompatible with '%s'", spec.ErrInvalidFilter, attr.Path())
}

// Parse the given raw value to the appropriate data type according to the type information in attribute.
// The attribute will be treated as singleValued even if it is multiValued.
func (t transformer) parseValue(raw string, attr *spec.Attribute) (interface{}, error) {
	if attr.Type() == spec.TypeComplex {
		return nil, fmt.Errorf("%w: operations cannot be applied to complex attribute", spec.ErrInvalidFilter)
	}
	switch attr.Type() {
	case spec.TypeString, spec.TypeReference, spec.TypeBinary:
		return unquote(raw), nil
	case spec.TypeDateTime:
		parsed, err := time.Parse(spec.ISO8601, unquote(raw))
		if err != nil {
			return nil, t.errIncompatibleValue(attr)
		}
		return primitive.NewDateTimeFromTime(parsed), nil
	case spec.TypeBoolean:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, t.errIncompatibleValue(attr)
		}
		return b, nil
	case spec.TypeInteger:
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, t.errIncompatibleValue(attr)
		}
		return i, nil
	case spec.TypeDecimal:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, t.errIncompatibleValue(attr)
		}
		return f, nil
	default:
		panic("impossible type")
	}
}

func unquote(raw string) string {
	uq, err := strconv.Unquote(raw)
	if err != nil {
		return raw
	}
	return uq
}

// pre compiled criteria
var (
	existsCriteria      = bson.D{{Key: mongoExists, Value: true}}
	nullCriteria        = bson.D{{Key: mongoNe, Value: primitive.Null{}}}
	emptyStringCriteria = bson.D{{Key: mongoNe, Value: ""}}
	emptyObjectCriteria = bson.D{{Key: mongoNe, Value: bson.M{}}}
	emptyArrayCriteria  = bson.D{
		{Key: mongoNot, Value: bson.A{
			bson.D{{Key: mongoSize, Value: 0}},
		}},
	}
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
