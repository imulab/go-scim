package scim

import (
	"errors"
	"fmt"
	"github.com/imulab/go-scim/internal/expr"
)

// compilePath calls expr.CompilePath and wraps any error in standard SCIM errors.
func compilePath(path string) (*expr.Node, error) {
	head, err := expr.CompilePath(path)
	if err != nil {
		if errors.Is(err, expr.ErrPath) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidPath, err)
		} else if errors.Is(err, expr.ErrFilter) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidFilter, err)
		} else {
			return nil, fmt.Errorf("%w: %s", ErrInternal, err)
		}
	}
	return head, nil
}

// compileFilter calls expr.CompileFilter and wraps any error in standard SCIM errors.
func compileFilter(filter string) (*expr.Node, error) {
	root, err := expr.CompileFilter(filter)
	if err != nil {
		if errors.Is(err, expr.ErrPath) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidPath, err)
		} else if errors.Is(err, expr.ErrFilter) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidFilter, err)
		} else {
			return nil, fmt.Errorf("%w: %s", ErrInternal, err)
		}
	}
	return root, nil
}
