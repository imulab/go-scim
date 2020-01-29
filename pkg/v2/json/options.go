package json

import (
	"github.com/imulab/go-scim/pkg/v2/prop"
	"strings"
)

// Return options to include given attributes in JSON serialization. Supplied attributes are still
// subject to SCIM rules for return-ability.
func Include(attributes ...string) Options {
	return include{attributes: attributes}
}

// Return options to exclude given attributes in JSON serialization. Supplied attributes are still
// subject to SCIM rules for return-ability.
func Exclude(attributes ...string) Options {
	return exclude{attributes: attributes}
}

type Options interface {
	apply(s *serializer, resource *prop.Resource)
}

type include struct {
	attributes []string
}

func (i include) apply(s *serializer, resource *prop.Resource) {
	if s.includes == nil {
		s.includes = []string{}
	}
	for _, path := range i.attributes {
		if len(path) > 0 {
			s.includes = append(s.includes, strings.TrimPrefix(
				strings.ToLower(path),
				strings.ToLower(resource.ResourceType().Schema().ID()+":")),
			)
		}
	}
}

type exclude struct {
	attributes []string
}

func (e exclude) apply(s *serializer, resource *prop.Resource) {
	if s.excludes == nil {
		s.excludes = []string{}
	}
	for _, path := range e.attributes {
		if len(path) > 0 {
			s.excludes = append(s.excludes, strings.TrimPrefix(
				strings.ToLower(path),
				strings.ToLower(resource.ResourceType().Schema().ID()+":")),
			)
		}
	}
}
