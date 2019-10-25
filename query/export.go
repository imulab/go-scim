package query

import "github.com/imulab/go-scim/core"

// Compile the given SCIM path and return the head of the path's step chain. The path is allowed
// to contain intermediate filter.
func CompilePath(path string, allowFilter bool) (*core.Step, error) {
	compiler := &pathCompiler{
		scan: &pathScanner{},
		data: append(copyOf(path), 0, 0),
		off:  0,
		op:   scanPathContinue,
	}
	compiler.scan.init()
	compiler.scan.allowFilter = allowFilter

	head := &core.Step{}
	cursor := head

	for compiler.hasMore() {
		next, err := compiler.next()
		if err != nil {
			return nil, err
		}
		cursor.Next = next
		cursor = cursor.Next
	}

	return head.Next, nil
}

// Compile the given SCIM filter and return the root of the filter's tree. The path within the filter
// is not allowed to contain additional filters.
func CompileFilter(query string) (*core.Step, error) {
	compiler := &filterCompiler{
		scan:    &filterScanner{},
		data:    append(copyOf(query), 0, 0),
		off:     0,
		op:      scanFilterSkipSpace,
		opStack: make([]*core.Step, 0),
		rsStack: make([]*core.Step, 0),
	}
	compiler.scan.init()

	for compiler.hasMore() {
		step, err := compiler.next()
		if err != nil {
			return nil, err
		}

		if step == nil {
			break
		}

		if step.IsLiteral() || step.IsPath() {
			if err := compiler.pushBuildResult(step); err != nil {
				return nil, err
			}
			continue
		}

		switch compiler.pushOperator(step) {
		case pushOpOk:
			break

		case pushOpRightParenthesis:
			for {
				popped := compiler.popOperatorIf(func(top *core.Step) bool {
					return !top.IsLeftParenthesis()
				})
				if popped != nil {
					// ignore error. we are sure it won't err
					_ = compiler.pushBuildResult(popped)
				} else {
					break
				}
			}
			if len(compiler.opStack) == 0 {
				return nil, core.Errors.InvalidFilter("mismatched parenthesis")
			} else {
				// discard the left parenthesis
				compiler.opStack = compiler.opStack[:len(compiler.opStack)-1]
			}
			break

		case pushOpInsufficientPriority:
			minPriority := opPriority(step.Token)
			for {
				popped := compiler.popOperatorIf(func(top *core.Step) bool {
					return opPriority(top.Token) >= minPriority
				})
				if popped != nil {
					// ignore error. we are sure it won't err
					_ = compiler.pushBuildResult(popped)
				} else {
					break
				}
			}
			if compiler.pushOperator(step) != pushOpOk {
				panic("flaw in algorithm")
			}
			break
		}
	}

	// pop all remaining operators
	for len(compiler.opStack) > 0 {
		_ = compiler.pushBuildResult(compiler.popOperatorIf(func(top *core.Step) bool {
			return true
		}))
	}

	// assertion check
	if len(compiler.rsStack) != 1 || !compiler.rsStack[0].IsOperator() {
		panic("flaw in algorithm")
	}

	// pop off the root so the rest could be GC'ed
	root := compiler.rsStack[0]
	compiler.rsStack = nil

	return root, nil
}

func copyOf(raw string) []byte {
	data := make([]byte, len(raw))
	copy(data, raw)
	return data
}

// Register recognized schema namespaces with this module, thus the compiler will recognize it as a
// separate path step.
func RegisterPathNamespace(ns string) {
	if len(ns) == 0 {
		return
	}
	namespaces = namespaces.insert(namespaces, ns, 0)
}
