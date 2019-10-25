package query

var (
	namespaces = &trie{}
)

type trie struct {
	// true if a word's trie path ends at this node
	w    bool
	next map[byte]*trie
}

func (t *trie) isWord() bool {
	return t != nil && t.w
}

func (t *trie) nextTrie(c byte) (*trie, bool) {
	if t == nil || len(t.next) == 0 {
		return nil, false
	}
	next, ok := t.next[t.toLowerCaseByte(c)]
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

	if x.next == nil {
		x.next = make(map[byte]*trie)
	}

	b := t.toLowerCaseByte(word[d])
	x.next[b] = t.insert(x.next[b], word, d+1)
	return x
}

func (_ *trie) toLowerCaseByte(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return 'a' + (c - 'A')
	}
	return c
}
