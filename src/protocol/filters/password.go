package filters

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	"golang.org/x/crypto/bcrypt"
)

// Return a field filter that BCrypt hashes the password field. It only performs the hashing when context does not have
// a PasswordBCryptedKey key. After hashing, it puts PasswordBCryptedKey into context to indicate that password is already
// hashed.
func NewPasswordFieldFilter(bcryptCost int, order int) protocol.FieldFilter {
	return &passwordFilter{
		cost:  bcryptCost,
		order: order,
	}
}

type (
	// Key in context to mark that the password value is already bcrypted.
	PasswordBCryptedKey struct{}
	passwordFilter      struct {
		order int
		cost  int
	}
)

func (f *passwordFilter) Supports(attribute *core.Attribute) bool {
	return attribute.ID() == "urn:ietf:params:scim:schemas:core:2.0:User:password"
}

func (f *passwordFilter) Order() int {
	return f.order
}

func (f *passwordFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property) error {
	return f.tryBCrypt(ctx, property)
}

func (f *passwordFilter) FieldRef(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	return f.tryBCrypt(ctx, property)
}

func (f *passwordFilter) tryBCrypt(ctx *protocol.FilterContext, property core.Property) error {
	if property.IsUnassigned() {
		return nil
	}

	_, ok := ctx.Get(PasswordBCryptedKey{})
	if ok {
		return nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(property.Raw().(string)), f.cost)
	if err != nil {
		return errors.Internal("failed to hash password: %s", err.Error())
	}

	_, err = prop.Internal(property).Replace(string(hashed))
	if err != nil {
		return err
	}

	ctx.Put(PasswordBCryptedKey{}, true)

	return nil
}
