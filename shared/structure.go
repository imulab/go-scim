package shared

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// SCIM resource
type Resource struct {
	Complex
}

func (r *Resource) GetId() string {
	if id, ok := r.Complex["id"].(string); ok {
		return id
	}
	return ""
}

func (r *Resource) GetData() Complex {
	return r.Complex
}

func ParseResource(filePath string) (*Resource, string, error) {
	path, err := filepath.Abs(filePath)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, "", err
	}

	data := make(map[string]interface{}, 0)
	err = json.Unmarshal(fileBytes, &data)
	if err != nil {
		return nil, "", err
	}

	return &Resource{Complex(data)}, string(fileBytes), nil
}

// SCIM complex data structure, Not thread-safe
type Complex map[string]interface{}

func (c Complex) Get(p Path, guide AttributeSource) chan interface{} {
	output := make(chan interface{})
	go func() {
		c.get(p, guide, output)
		close(output)
	}()
	return output
}

func (c Complex) get(p Path, guide AttributeSource, output chan interface{}) {
	attr := guide.GetAttribute(p, false)
	if attr == nil {
		return
	}

	if v, ok := c[attr.Name]; ok && v != nil {
		if p.FilterRoot() != nil {
			if mv, ok := v.([]interface{}); ok && mv != nil {
				matches := MultiValued(mv).Filter(p.FilterRoot(), attr)
				for match := range matches {
					if p.Next() != nil {
						if v0, ok := match.(map[string]interface{}); ok && v0 != nil {
							Complex(v0).get(p.Next(), attr, output)
						}
					} else {
						output <- match
					}
				}
			}
		} else {
			if p.Next() != nil {
				if v0, ok := v.(map[string]interface{}); ok && v0 != nil {
					Complex(v0).get(p.Next(), attr, output)
				}
			} else {
				output <- v
			}
		}
	}
}

// Evaluate given predicate
func (c Complex) Evaluate(filter FilterNode, guide AttributeSource) bool {
	return newPredicate(filter, guide).evaluate(c)
}

// Set the value at the specified Path
// Path is a dot separated string that may contain filter
func (c Complex) Set(p Path, value interface{}, guide AttributeSource) error {
	attr := guide.GetAttribute(p, true)
	if attr == nil {
		return Error.InvalidPath(p.CollectValue(), "no attribute found")
	}

	// TODO validate

	base, last := p.SeparateAtLast()
	itemsToSet := make(chan interface{})
	go func() {
		if base != nil {
			c.get(base, guide, itemsToSet)
		} else {
			itemsToSet <- c
		}
		close(itemsToSet)
	}()

	for item := range itemsToSet {
		if m, ok := item.(Complex); ok && m != nil {
			m.set(last, value, attr)
		} else if m, ok := item.(map[string]interface{}); ok && m != nil {
			Complex(m).set(last, value, attr)
		}
	}
	return nil
}

func (c Complex) set(p Path, value interface{}, attr *Attribute) {
	if attr.MultiValued && p.FilterRoot() != nil {
		if mv, ok := c[attr.Name].([]interface{}); ok {
			for i, v := range mv {
				if c0, ok := v.(map[string]interface{}); ok {
					if Complex(c0).Evaluate(p.FilterRoot(), attr) {
						MultiValued(mv).Set(i, value)
					}
				}
			}
		}
	} else {
		c[attr.Name] = value
	}
}

// SCIM multivalued data structure, Not thread-safe
type MultiValued []interface{}

func (c MultiValued) Get(index int) interface{} {
	return c[index]
}

func (c MultiValued) Set(index int, value interface{}) {
	c[index] = value
}

func (c MultiValued) Len() int {
	return len(c)
}

func (c MultiValued) Add(value ...interface{}) MultiValued {
	return MultiValued(append([]interface{}(c), value...))
}

func (c MultiValued) Remove(index int) MultiValued {
	// TODO
	return nil
}

func (c MultiValued) Filter(root FilterNode, guide AttributeSource) chan interface{} {
	output := make(chan interface{})
	go func() {
		for _, elem := range c {
			if m, ok := elem.(map[string]interface{}); ok {
				if Complex(m).Evaluate(root, guide) {
					output <- elem
				}
			}
		}
		close(output)
	}()
	return output
}
