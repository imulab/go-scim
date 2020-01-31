package json

import (
	"strings"
)

// Include returns Options to include given attributes in JSON serialization. Supplied attributes are still
// subject to SCIM rules for return-ability.
func Include(attributes ...string) Options {
	return include{attributes: attributes}
}

// Exclude returns Options to exclude given attributes in JSON serialization. Supplied attributes are still
// subject to SCIM rules for return-ability.
func Exclude(attributes ...string) Options {
	return exclude{attributes: attributes}
}

// JSON serialization options.
type Options interface {
	apply(s *serializer, serializable Serializable)
}

type include struct {
	attributes []string
}

func (i include) apply(s *serializer, serializable Serializable) {
	if s.includes == nil {
		s.includes = []string{}
	}
	for _, path := range i.attributes {
		if len(path) > 0 {
			s.includes = append(s.includes, strings.TrimPrefix(
				strings.ToLower(path),
				strings.ToLower(serializable.MainSchemaId()+":")),
			)
		}
	}
}

type exclude struct {
	attributes []string
}

func (e exclude) apply(s *serializer, serializable Serializable) {
	if s.excludes == nil {
		s.excludes = []string{}
	}
	for _, path := range e.attributes {
		if len(path) > 0 {
			s.excludes = append(s.excludes, strings.TrimPrefix(
				strings.ToLower(path),
				strings.ToLower(serializable.MainSchemaId()+":")),
			)
		}
	}
}
