package expr

import "strconv"

// isFirstAlphabet returns true if the byte can be the first alphabet of a SCIM attribute name.
func isFirstAlphabet(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '$'
}

// isNonFirstAlphabet returns true if the byte can be the non-first alphabet of a SCIM attribute name.
func isNonFirstAlphabet(c byte) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9') ||
		c == '-' ||
		c == '_'
}

func quoteChar(c byte) string {
	// special cases - different from quoted strings
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}

	// use quoted string with different quotation marks
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}

func copyOf(raw string) []byte {
	data := make([]byte, len(raw))
	copy(data, raw)
	return data
}
