package expr

import "github.com/imulab/go-scim/src/core"

var urnsCache = &urns{}

// Register the resource type to correctly use expression package's compiler capability. This method
// caches all schema urn ids available in a resource type, so they can be recognized later when an
// expression that contains one is passed in as an argument to compiler.
func Register(resourceType *core.ResourceType) {
	register(resourceType.Schema().ID())
	resourceType.ForEachExtension(func(extension *core.Schema, _ bool) {
		register(extension.ID())
	})
}

func register(s string) {
	urnsCache = urnsCache.insert(urnsCache, s, 0)
}

// A trie data structure to cache all registered resource type ID URNs. These URNs
// are essential for the compiler to decide where to treat a dot as a path separator
// and where to treat it as just part of the property namespace.
type urns struct {
	// true if a word's trie path ends at this node
	w    bool
	next map[byte]*urns
}

func (t *urns) isWord() bool {
	return t != nil && t.w
}

func (t *urns) nextTrie(c byte) (*urns, bool) {
	if t == nil || len(t.next) == 0 {
		return nil, false
	}
	next, ok := t.next[toLowerCaseByte(c)]
	return next, ok
}

func (t *urns) insert(x *urns, word string, d int) *urns {
	if x == nil {
		x = &urns{}
	}

	if d == len(word) {
		x.w = true
		return x
	}

	if x.next == nil {
		x.next = make(map[byte]*urns)
	}

	b := toLowerCaseByte(word[d])
	x.next[b] = t.insert(x.next[b], word, d+1)
	return x
}

func toLowerCaseByte(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return 'a' + (c - 'A')
	}
	return c
}
