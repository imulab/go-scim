package shared

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

func ValidateUniqueness(subj *Resource, sch *Schema, repo Repository, ctx context.Context) (err error) {
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

	uniquenessValidatorInstance.validateUniquenessWithReflection(reflect.ValueOf(subj.Complex), sch.ToAttribute(), repo, ctx)
	return
}

var (
	oneUniquenessValidator      sync.Once
	uniquenessValidatorInstance *uniquenessValidator
)

func init() {
	oneUniquenessValidator.Do(func() {
		uniquenessValidatorInstance = &uniquenessValidator{}
	})
}

type uniquenessValidator struct{}

func (uv *uniquenessValidator) validateUniquenessWithReflection(v reflect.Value, guide *Attribute, repo Repository, ctx context.Context) {
	for _, attr := range guide.SubAttributes {
		v0 := v.MapIndex(reflect.ValueOf(attr.Name))
		if !attr.Assigned(v0) {
			continue
		}
		if v0.Kind() == reflect.Interface {
			v0 = v0.Elem()
		}

		switch attr.Uniqueness {
		case Server, Global:
			query := fmt.Sprintf("%s eq \"%v\"", attr.Assist.Path, v0.Interface())
			count, err := repo.Count(query, ctx)
			if err != nil {
				uv.throw(err, ctx)
			} else if count > 0 {
				requestType := ctx.Value(RequestType{}).(int)
				switch requestType {
				case ReplaceUser, ReplaceGroup, PatchUser, PatchGroup:
					if count > 1 {
						uv.throw(Error.Duplicate(attr.Assist.Path, v0.Interface()), ctx)
					} else {
						resourceId := ctx.Value(ResourceId{}).(string)
						lr, err := repo.Search(SearchRequest{Filter: query, StartIndex: 1}, ctx)
						if err != nil {
							uv.throw(Error.Text("Cannot verify uniqueness: %s", err.Error()), ctx)
						}
						if resourceId != lr.Resources[0].GetData()["id"].(string) {
							uv.throw(Error.Duplicate(attr.Assist.Path, v0.Interface()), ctx)
						}
					}
				default:
					uv.throw(Error.Duplicate(attr.Assist.Path, v0.Interface()), ctx)
				}
			}
		}

		if attr.ExpectsComplex() && v0.Kind() == reflect.Map {
			uv.validateUniquenessWithReflection(v0, attr, repo, ctx)
		}
	}
}

func (uv *uniquenessValidator) throw(err error, ctx context.Context) {
	panic(err)
}
