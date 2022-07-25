package scim

type trie struct {
	w bool
	m map[byte]*trie
}

func (t *trie) isWord() bool {
	return t != nil && t.w
}

func (t *trie) nextTrie(c byte) (*trie, bool) {
	if t == nil || len(t.m) == 0 {
		return nil, false
	}

	next, ok := t.m[toLowerCaseByte(c)]
	return next, ok
}

func (t *trie) insert(x *trie, word string, d int) *trie {
	if x == nil {
		x = &trie{}
	}

	if d == len(word) {
		x.w = true
		return x
	}

	if x.m == nil {
		x.m = make(map[byte]*trie)
	}

	b := toLowerCaseByte(word[d])
	x.m[b] = t.insert(x.m[b], word, d+1)

	return x
}

func toLowerCaseByte(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return 'a' + (c - 'A')
	}
	return c
}
