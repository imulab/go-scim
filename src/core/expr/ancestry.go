package expr

import (
	"github.com/imulab/go-scim/src/core"
	"strings"
)

type (
	// The path ancestry is a trie data structure that answers one essential question:
	// Is the querying path among the registered paths, or a ancestor of one or more
	// registered paths, or a potential offspring of a registered path?
	// For instance:
	// - when 'emails.value' is registered, 'emails' is an ancestor of this family.
	// - when 'emails' is registered, 'emails.value' is a potential offspring of this family.
	//
	// This functionality will help the feature of returned-ability. When client includes a path
	// in the 'attributes' or 'excludedAttributes' parameter, we not only need to match that
	// path, and also need to infer on its ancestor path and offspring path.
	//
	// In general, when part of the included 'attributes', both ancestor and offspring of a path
	// need to also be included. When part of the 'excludedAttributes', any offspring of the path
	// is also excluded, but not their ancestors.
	PathAncestry struct {
		namespace string
		root      *pathGeneration
	}
	// A single generation within the ancestry
	pathGeneration struct {
		hasPath		bool
		offspring 	map[string]*pathGeneration
	}
)

func NewPathFamily(resourceType *core.ResourceType) *PathAncestry {
	return &PathAncestry{
		namespace: resourceType.Schema().ID(),
		root:      &pathGeneration{},
	}
}

// Add a path expression to the family.
func (p *PathAncestry) Add(path *Expression) {
	if path == nil {
		return
	}

	if strings.ToLower(path.Token()) == strings.ToLower(p.namespace) {
		p.Add(path.next)
		return
	}

	p.root = p.root.add(p.root, path)
}

func (g *pathGeneration) add(x *pathGeneration, path *Expression) *pathGeneration {
	if x == nil {
		x = &pathGeneration{}
	}

	if path == nil {
		x.hasPath = true
		return x
	}

	if x.offspring == nil {
		x.offspring = make(map[string]*pathGeneration)
	}

	x.offspring[strings.ToLower(path.Token())] = g.add(x.offspring[strings.ToLower(path.Token())], path.next)
	return x
}

// Returns true if the path denoted by the expression list is a member of this
// path family. A path becomes a member only when it is previously added to the
// family using Add method.
func (p *PathAncestry) IsMember(path *Expression) bool {
	if path == nil || path.ContainsFilter() {
		return false
	}

	if strings.ToLower(path.Token()) == strings.ToLower(p.namespace) {
		return p.IsMember(path.next)
	}

	return p.root.isMember(path)
}

func (g *pathGeneration) isMember(path *Expression) bool {
	if path == nil {
		return g.hasPath
	}

	if len(g.offspring) == 0 {
		return false
	}

	if k, ok := g.offspring[strings.ToLower(path.Token())]; !ok {
		return false
	} else {
		return k.isMember(path.next)
	}
}

// Returns true if the path denoted by the expression is an ancestor of any path that
// is a member of this family. For instance, if the members of this family are 'emails.value',
// 'emails.primary' and 'userName', then 'emails' will be an ancestor of this family, but
// 'name' will not.
func (p *PathAncestry) IsAncestor(path *Expression) bool {
	if path == nil || path.ContainsFilter() {
		return false
	}

	if strings.ToLower(path.Token()) == strings.ToLower(p.namespace) {
		return p.IsAncestor(path.next)
	}

	if len(p.root.offspring) == 0 {
		return false
	}

	return p.root.isAncestor(path)
}

func (g *pathGeneration) isAncestor(path *Expression) bool {
	if path == nil {
		return !g.hasPath
	}

	if len(g.offspring) == 0 {
		return false
	}

	if k, ok := g.offspring[strings.ToLower(path.Token())]; !ok {
		return false
	} else {
		return k.isAncestor(path.next)
	}
}

// Returns true if the path denoted by the expression if a potential offspring of a member of
// this family. For instance, if the members of this family are 'emails' and 'phoneNumbers', then
// 'emails.value' is an offspring, 'phoneNumbers.type' is an offspring, but 'name' or 'groups.$ref' is
// not an offspring.
func (p *PathAncestry) IsOffspring(path *Expression) bool {
	if path == nil || path.ContainsFilter() {
		return false
	}

	if strings.ToLower(path.Token()) == strings.ToLower(p.namespace) {
		return p.IsOffspring(path.next)
	}

	if len(p.root.offspring) == 0 {
		return false
	}

	return p.root.isOffspring(path)
}

func (g *pathGeneration) isOffspring(path *Expression) bool {
	if path == nil {
		return false
	}

	if len(g.offspring) == 0 {
		return true
	}

	if k, ok := g.offspring[strings.ToLower(path.Token())]; !ok {
		return true
	} else {
		return k.isOffspring(path.next)
	}
}