package expr

import "strings"

var (
	pathCache = make(map[string]*Expression)
	filterCache = make(map[string]*Expression)
)

// MustPath try to get a compiled path from cache, if does not exist, then compile it and cache it.
// This method is intended for internal usage. Do not call with user provided paths as the method
// panics on any error.
func MustPath(path string) *Expression {
	if p, ok := pathCache[strings.ToLower(path)]; ok {
		return p
	} else {
		compiled, err := CompilePath(path)
		if err != nil {
			panic(err)
		}
		pathCache[strings.ToLower(path)] = compiled
		return compiled
	}
}

// MustFilter try to get a compiled filter from cache, if does not exist, then compile it and cache it.
// This method is intended for internal usage. Do not call with user provided filters as the method
// panics on any error.
func MustFilter(filter string) *Expression {
	if f, ok := filterCache[strings.ToLower(filter)]; ok {
		return f
	} else {
		compiled, err := CompileFilter(filter)
		if err != nil {
			panic(err)
		}
		filterCache[strings.ToLower(filter)] = compiled
		return compiled
	}
}