package filter

import (
	"context"
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/core/prop"
	"golang.org/x/crypto/bcrypt"
)

const fieldPassword = "password"

// Return a filter that hashes the password field using BCrypt algorithm. It only performs the hashing when
// the value is not yet bcrypted.
func Password(bcryptCost int) ForResource {
	return &passwordFilter{
		cost: bcryptCost,
	}
}

type passwordFilter struct {
	cost int
}

func (f *passwordFilter) Filter(ctx context.Context, resource *prop.Resource) error {
	return f.tryBCrypt(resource)
}

func (f *passwordFilter) FilterRef(ctx context.Context, resource *prop.Resource, ref *prop.Resource) error {
	return f.tryBCrypt(resource)
}

func (f *passwordFilter) tryBCrypt(resource *prop.Resource) error {
	pwdProp, err := resource.NewNavigator().FocusName(fieldPassword)
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
