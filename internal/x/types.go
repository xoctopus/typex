package x

import (
	"reflect"
)

type Type interface {
	// Unwrap to types.Type or reflect.Type
	Unwrap() any

	Kind() reflect.Kind

	PkgPath() string
	Name() string
	String() string
	TypeLit() string

	Implements(any) bool
	AssignableTo(any) bool
	ConvertibleTo(any) bool
	Comparable() bool

	Key() Type
	Elem() Type
	Len() int

	NumField() int
	Field(int) StructField
	FieldByName(string) (StructField, bool)
	FieldByNameFunc(func(string) bool) (StructField, bool)

	NumMethod() int
	Method(int) Method
	MethodByName(string) (Method, bool)

	IsVariadic() bool
	NumIn() int
	In(int) Type
	NumOut() int
	Out(int) Type
}

type Method interface {
	PkgPath() string
	Name() string
	Type() Type
}

type StructField interface {
	PkgPath() string
	Name() string
	Type() Type
	Tag() reflect.StructTag
	Anonymous() bool
}
