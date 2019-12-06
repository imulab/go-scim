package filters

import (
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol"
	"golang.org/x/crypto/bcrypt"
)

// Return a field filter that BCrypt hashes the password field. It only performs the hashing when the value is not yet
// bcrypted.
func NewPasswordResourceFilter(bcryptCost int) protocol.ResourceFilter {
	return &passwordFilter{
		cost: bcryptCost,
	}
}

type passwordFilter struct {
	cost int
}

func (f *passwordFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource) error {
	return f.tryBCrypt(ctx, resource)
}

func (f *passwordFilter) FilterRef(ctx *protocol.FilterContext, resource *prop.Resource, ref *prop.Resource) error {
	return f.tryBCrypt(ctx, resource)
}

func (f *passwordFilter) tryBCrypt(ctx *protocol.FilterContext, resource *prop.Resource) error {
	pwdProp, err := resource.NewNavigator().FocusName("password")
	if err != nil {
		return err
	}

	if pwdProp.IsUnassigned() {
		return nil
	}

	if _, err := bcrypt.Cost([]byte(pwdProp.Raw().(string))); err == nil {
		// field is already bcrypted
		return nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(pwdProp.Raw().(string)), f.cost)
	if err != nil {
		return errors.Internal("failed to hash password: %s", err.Error())
	}

	err = pwdProp.Replace(string(hashed))
	if err != nil {
		return err
	}

	return nil
}
