package log

import (
	"fmt"
	"strings"
)

type Args map[string]interface{}

func (a Args) Count() int {
	return len(a)
}

func (a Args) ForEach(f func(k string, v interface{})) {
	for k, v := range a {
		f(k, v)
	}
}

func (a Args) String() string {
	s := make([]string, 0)
	a.ForEach(func(k string, v interface{}) {
		s = append(s, fmt.Sprintf("%s=%v", k, v))
	})
	return strings.Join(s, " ")
}