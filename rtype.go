package typex

import (
	"go/types"
	"reflect"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"

	"github.com/xoctopus/typex/internal"
)

func NewRType(t reflect.Type) Type {
	must.NotNilF(t, "invalid reflect.Type")
	return &rtype{t: t, u: internal.Global().Literalize(t)}
}

type rtype struct {
	t reflect.Type
	u internal.Literal
}

func (t *rtype) Unwrap() any { return t.t }

func (t *rtype) Kind() reflect.Kind { return t.t.Kind() }

func (t *rtype) PkgPath() string { return t.t.PkgPath() }

func (t *rtype) Name() string { return t.t.Name() }

func (t *rtype) String() string { return t.u.String() }

func (t *rtype) Alias() string {
	if x, ok := t.u.(internal.Builtin); ok {
		return x.Alias()
	}
	return ""
}

func (t *rtype) TypeLit() string { return t.u.TypeLit() }

func (t *rtype) Implements(u any) bool {
	switch x := u.(type) {
	case Type:
		return t.Implements(x.Unwrap())
	case reflect.Type:
		if x.Kind() == reflect.Interface {
			return t.t.Implements(x)
		}
		return false
	case types.Type:
		if i, ok := x.Underlying().(*types.Interface); ok {
			return types.Implements(internal.Global().TType(t.t), i)
		}
		return false
	default:
		return false
	}
}

func (t *rtype) AssignableTo(u any) bool {
	switch x := u.(type) {
	case Type:
		return t.AssignableTo(x.Unwrap())
	case reflect.Type:
		return t.t.AssignableTo(x)
	case types.Type:
		return types.AssignableTo(internal.Global().TType(t.t), x)
	default:
		return false
	}
}

func (t *rtype) ConvertibleTo(u any) bool {
	switch x := u.(type) {
	case Type:
		return t.ConvertibleTo(x.Unwrap())
	case reflect.Type:
		return t.t.ConvertibleTo(x)
	case types.Type:
		return types.ConvertibleTo(internal.Global().TType(t.t), x)
	default:
		return false
	}
}

func (t *rtype) Comparable() bool { return t.t.Comparable() }

func (t *rtype) Key() Type {
	if t.Kind() == reflect.Map {
		return NewRType(t.t.Key())
	}
	return nil
}

func (t *rtype) Elem() Type {
	if reflectx.CanElem(t.t) {
		return NewRType(t.t.Elem())
	}
	return nil
}

func (t *rtype) Len() int {
	if t.t.Kind() == reflect.Array {
		return t.t.Len()
	}
	return 0
}

func (t *rtype) NumField() int {
	if t.Kind() == reflect.Struct {
		return t.t.NumField()
	}
	return 0
}

func (t *rtype) Field(i int) StructField {
	if i >= 0 && i < t.NumField() {
		return &RStructField{StructField: t.t.Field(i)}
	}
	return nil
}

func (t *rtype) FieldByName(name string) (StructField, bool) {
	if t.Kind() == reflect.Struct {
		if f, ok := t.t.FieldByName(name); ok {
			return &RStructField{f}, true
		}
	}
	return nil, false
}

func (t *rtype) FieldByNameFunc(match func(string) bool) (StructField, bool) {
	if t.Kind() == reflect.Struct {
		if f, ok := t.t.FieldByNameFunc(match); ok {
			return &RStructField{f}, true
		}
	}
	return nil, false
}

func (t *rtype) NumMethod() int {
	return t.t.NumMethod()
}

func (t *rtype) Method(i int) Method {
	if i >= 0 && i < t.NumMethod() {
		return &RMethod{t.t.Method(i)}
	}
	return nil
}

func (t *rtype) MethodByName(name string) (Method, bool) {
	if m, ok := t.t.MethodByName(name); ok {
		return &RMethod{m}, true
	}
	return nil, false
}

func (t *rtype) IsVariadic() bool {
	if t.Kind() == reflect.Func {
		return t.t.IsVariadic()
	}
	return false
}

func (t *rtype) NumIn() int {
	if t.Kind() == reflect.Func {
		return t.t.NumIn()
	}
	return 0
}

func (t *rtype) In(i int) Type {
	if t.Kind() == reflect.Func && i >= 0 && i < t.t.NumIn() {
		return NewRType(t.t.In(i))
	}
	return nil
}

func (t *rtype) NumOut() int {
	if t.Kind() == reflect.Func {
		return t.t.NumOut()
	}
	return 0
}

func (t *rtype) Out(i int) Type {
	if t.Kind() == reflect.Func && i >= 0 && i < t.t.NumOut() {
		return NewRType(t.t.Out(i))
	}
	return nil
}

type RStructField struct {
	reflect.StructField
}

func (f *RStructField) PkgPath() string {
	return f.StructField.PkgPath
}

func (f *RStructField) Name() string {
	return f.StructField.Name
}

func (f *RStructField) Type() Type {
	return NewRType(f.StructField.Type)
}

func (f *RStructField) Tag() reflect.StructTag {
	return f.StructField.Tag
}

func (f *RStructField) Anonymous() bool {
	return f.StructField.Anonymous
}

type RMethod struct {
	reflect.Method
}

func (m *RMethod) PkgPath() string {
	return m.Method.PkgPath
}

func (m *RMethod) Name() string {
	return m.Method.Name
}

func (m *RMethod) Type() Type {
	return NewRType(m.Method.Type)
}
