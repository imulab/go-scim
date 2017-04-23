package shared

import (
	"context"
	"reflect"
)

func ApplyPatch(patch Patch, subj *Resource, sch *Schema, ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			default:
				err = Error.Text("%v", r)
			}
		}
	}()

	ps := patchState{patch:patch}

	var path Path
	if len(patch.Path) == 0 {
		path = nil
	} else {
		path, err = NewPath(patch.Path)
		if err != nil {
			return err
		}
		path.CorrectCase(sch, true)

		if attr := sch.GetAttribute(path, true); attr != nil {
			ps.destAttr = attr
		} else {
			return Error.InvalidPath(patch.Path, "no attribute found for path")
		}
	}

	v := reflect.ValueOf(patch.Value)
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch patch.Op {
	case Add:
		ps.applyPatchAdd(path, v, subj, sch, ctx)
	case Replace:
		ps.applyPatchReplace(path, v, subj, sch, ctx)
	case Remove:
		ps.applyPatchRemove(path, subj, sch, ctx)
	default:
		err = Error.InvalidParam("Op", "one of [add|remove|replace]", patch.Op)
	}
	return
}

type patchState struct {
	patch 		Patch
	destAttr 	*Attribute
}

func (ps *patchState) throw(err error, ctx context.Context) {
	if err != nil {
		panic(err)
	}
}

func (ps *patchState) applyPatchRemove(p Path, subj *Resource, sch *Schema, ctx context.Context) {
	basePath, lastPath := p.SeparateAtLast()
	baseChannel := make(chan interface{}, 1)
	if basePath == nil {
		go func(){
			baseChannel <- subj.Complex
			close(baseChannel)
		}()
	} else {
		baseChannel = subj.Get(basePath, sch)
	}

	var baseAttr AttributeSource = sch
	if basePath != nil {
		baseAttr = sch.GetAttribute(basePath, true)
	}

	for base := range baseChannel {
		baseVal := reflect.ValueOf(base)
		if baseVal.IsNil() {
			continue
		}
		if baseVal.Kind() == reflect.Interface {
			baseVal = baseVal.Elem()
		}

		switch baseVal.Kind() {
		case reflect.Map:
			keyVal := reflect.ValueOf(lastPath.Base())
			if ps.destAttr.MultiValued {
				if lastPath.FilterRoot() == nil {
					baseVal.SetMapIndex(keyVal, reflect.Value{})
				} else {
					origVal := baseVal.MapIndex(keyVal)
					baseAttr = baseAttr.GetAttribute(lastPath, false)
					reverseRoot := &filterNode{
						data:Not,
						typ:LogicalOperator,
						left:lastPath.FilterRoot().(*filterNode),
					}
					newElemChannel := MultiValued(origVal.Interface().([]interface{})).Filter(reverseRoot, baseAttr)
					newArr := make([]interface{}, 0)
					for newElem := range newElemChannel {
						newArr = append(newArr, newElem)
					}
					if len(newArr) == 0 {
						baseVal.SetMapIndex(keyVal, reflect.Value{})
					} else {
						baseVal.SetMapIndex(keyVal, reflect.ValueOf(newArr))
					}
				}
			} else {
				baseVal.SetMapIndex(keyVal, reflect.Value{})
			}
		case reflect.Array, reflect.Slice:
			keyVal := reflect.ValueOf(lastPath.Base())
			for i := 0; i < baseVal.Len(); i++ {
				elemVal := baseVal.Index(i)
				if elemVal.Kind() == reflect.Interface {
					elemVal = elemVal.Elem()
				}
				switch elemVal.Kind() {
				case reflect.Map:
					elemVal.SetMapIndex(keyVal, reflect.Value{})
				default:
					ps.throw(Error.InvalidPath(ps.patch.Path, "array base contains non-map"), ctx)
				}
			}
		default:
			ps.throw(Error.InvalidPath(ps.patch.Path, "base evaluated to non-map and non-array."), ctx)
		}
	}
}

func (ps *patchState) applyPatchReplace(p Path, v reflect.Value, subj *Resource, sch *Schema, ctx context.Context) {
	basePath, lastPath := p.SeparateAtLast()
	baseChannel := make(chan interface{}, 1)
	if basePath == nil {
		go func(){
			baseChannel <- subj.Complex
			close(baseChannel)
		}()
	} else {
		baseChannel = subj.Get(basePath, sch)
	}

	for base := range baseChannel {
		baseVal := reflect.ValueOf(base)
		if baseVal.IsNil() {
			continue
		}
		if baseVal.Kind() == reflect.Interface {
			baseVal = baseVal.Elem()
		}
		baseVal.SetMapIndex(reflect.ValueOf(lastPath.Base()), v)
	}
}

func (ps *patchState) applyPatchAdd(p Path, v reflect.Value, subj *Resource, sch *Schema, ctx context.Context) {
	if p == nil {
		if v.Kind() != reflect.Map {
			ps.throw(Error.InvalidParam("value of add op", "to be complex (for implicit path)", "non-complex"), ctx)
		}
		for _, k := range v.MapKeys() {
			v0 := v.MapIndex(k)
			if err := ApplyPatch(Patch{
				Op:Add,
				Path:k.String(),
				Value:v0.Interface(),
			}, subj, sch, ctx); err != nil {
				ps.throw(err, ctx)
			}
		}
	} else {
		basePath, lastPath := p.SeparateAtLast()
		baseChannel := make(chan interface{}, 1)

		if basePath == nil {
			go func(){
				baseChannel <- subj.Complex
				close(baseChannel)
			}()
		} else {
			baseChannel = subj.Get(basePath, sch)
		}

		for base := range baseChannel {
			baseVal := reflect.ValueOf(base)
			if baseVal.IsNil() {
				continue
			}
			if baseVal.Kind() == reflect.Interface {
				baseVal = baseVal.Elem()
			}

			switch baseVal.Kind() {
			case reflect.Map:
				keyVal := reflect.ValueOf(lastPath.Base())
				if ps.destAttr.MultiValued {
					origVal := baseVal.MapIndex(keyVal)
					if !origVal.IsValid() {
						baseVal.SetMapIndex(keyVal, reflect.ValueOf([]interface{}{v.Interface()}))
					} else {
						if origVal.Kind() == reflect.Interface {
							origVal = origVal.Elem()
						}
						newArr := MultiValued(origVal.Interface().([]interface{})).Add(v.Interface())
						baseVal.SetMapIndex(keyVal, reflect.ValueOf(newArr))
					}
				} else {
					baseVal.SetMapIndex(keyVal, v)
				}
			case reflect.Array, reflect.Slice:
				for i := 0; i < baseVal.Len(); i++ {
					elemVal := baseVal.Index(i)
					if elemVal.Kind() == reflect.Interface {
						elemVal = elemVal.Elem()
					}
					switch elemVal.Kind() {
					case reflect.Map:
						elemVal.SetMapIndex(reflect.ValueOf(lastPath.Base()), v)
					default:
						ps.throw(Error.InvalidPath(ps.patch.Path, "array base contains non-map"), ctx)
					}
				}
			default:
				ps.throw(Error.InvalidPath(ps.patch.Path, "base evaluated to non-map and non-array."), ctx)
			}
		}
	}
}

