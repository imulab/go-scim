package expr

import "errors"

var (
	ErrFilter = errors.New("failed to compile filter")
	ErrPath   = errors.New("failed to compile path")
)
