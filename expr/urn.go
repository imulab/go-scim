package expr

import "sync"

var (
	urnLookupLock sync.RWMutex
	urnLookup     *byteTrie
)

// RegisterURN registers the given schema id URN with the global registry. This method is protected by a RW lock and
// thus concurrent-safe. Nonetheless, it is expected that this function be called at initialization and keep the registry
// read only afterwards.
func RegisterURN(urn string) {
	if len(urn) == 0 {
		return
	}

	urnLookupLock.Lock()
	urnLookup = insertWord(urnLookup, urn, 0)
	urnLookupLock.Unlock()
}

// byteTrie is a Trie data structure that hosts a single byte or character. It is purposed to record SCIM schema id
// URN prefixes which could contain the dot (".") character, which would otherwise be mistaken to be the SCIM path
// separator. If a path name start with characters that can be matched in this trie and then a semi-colon (":", separator
// used to delimit URN prefixes and actual path names), the dot (".") within shall be treated as part of the URN prefix,
// instead of a path separator.
type byteTrie struct {
	word      bool
	nextBytes map[byte]*byteTrie
}

func (t *byteTrie) isWord() bool {
	return t != nil && t.word
}

// next returns the remaining trie structure if the given byte is matched in the current level, or nil if no match.
func (t *byteTrie) next(c byte) *byteTrie {
	if t == nil || len(t.nextBytes) == 0 {
		return nil
	}
	return t.nextBytes[toLowerCaseByte(c)]
}

func toLowerCaseByte(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return 'a' + (c - 'A')
	}
	return c
}

// insertWord inserts the given word into the byteTrie. It does so by putting the character at the given depth into
// the current root and recurse to depth plus one.
func insertWord(root *byteTrie, word string, depth int) *byteTrie {
	if root == nil {
		root = &byteTrie{}
	}

	if depth == len(word) {
		root.word = true
		return root
	}

	if root.nextBytes == nil {
		root.nextBytes = make(map[byte]*byteTrie)
	}

	b := toLowerCaseByte(word[depth])
	root.nextBytes[b] = insertWord(root.nextBytes[b], word, depth+1)

	return root
}
