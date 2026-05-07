package filter

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

// BCryptFilter returns a ByProperty filter that hashes data using the BCrypt algorithm for string or binary properties
// whose attribute is annotated with @BCrypt. If the property is unassigned or has the same value with the reference
// property, the filter does nothing. Otherwise, it will attempt to determine the cost through the "cost" annotation
// parameter and replace the property value with the hashed value. For binary properties specifically, the hashed value
// is base64 encoded before replacing the original base64 encoded bytes.
func BCryptFilter() ByProperty {
	return bCryptPropertyFilter{}
}

type bCryptPropertyFilter struct{}

func (f bCryptPropertyFilter) Supports(attribute *spec.Attribute) bool {
	if _, ok := attribute.Annotation(annotation.BCrypt); !ok {
		return false
	}
	return !attribute.MultiValued() && (attribute.Type() == spec.TypeString || attribute.Type() == spec.TypeBinary)
}

func (f bCryptPropertyFilter) Filter(_ context.Context, _ *spec.ResourceType, nav prop.Navigator) error {
	if nav.HasError() {
		return nav.Error()
	}

	if nav.Current().IsUnassigned() {
		return nil
	}

	return f.bCryptAndReplace(nav)
}

func (f bCryptPropertyFilter) FilterRef(_ context.Context, _ *spec.ResourceType, nav prop.Navigator, refNav prop.Navigator) error {
	if nav.HasError() {
		return nav.Error()
	}

	if nav.Current().IsUnassigned() {
		return nil
	}

	if refNav != nil && nav.Current().Raw() == refNav.Current().Raw() {
		// property value is the same as reference value. Reference value
		// are values from database that can be assumed to have undergone
		// this filter already. Values being the same indicates no additional
		// bCrypt is needed because it is not a new value.
		return nil
	}

	return f.bCryptAndReplace(nav)
}

// perform bCrypt on the current property and replace the property value locally. Property must not be unassigned and
// must only be type string or type binary.
func (f bCryptPropertyFilter) bCryptAndReplace(nav prop.Navigator) error {
	attr := nav.Current().Attribute()

	var raw []byte
	switch attr.Type() {
	case spec.TypeString:
		raw = []byte(nav.Current().Raw().(string))
	case spec.TypeBinary:
		raw, _ = base64.StdEncoding.DecodeString(nav.Current().Raw().(string))
	default:
		panic("unsupported type")
	}

	var cost int
	params, _ := attr.Annotation(annotation.BCrypt)
	cost, err := strconv.Atoi(fmt.Sprintf("%v", params["cost"]))
	if err != nil || cost < 1 {
		cost = 10
	}

	hashed, err := bcrypt.GenerateFromPassword(raw, cost)
	if err != nil {
		return fmt.Errorf("%w: failed to perform bCrypt on attribute '%s'", spec.ErrInternal, attr.Path())
	}

	var replacement string
	switch attr.Type() {
	case spec.TypeString:
		replacement = string(hashed)
	case spec.TypeBinary:
		replacement = base64.StdEncoding.EncodeToString(hashed)
	default:
		panic("unsupported type")
	}

	_, err = nav.Current().Replace(replacement)
	return err
}
