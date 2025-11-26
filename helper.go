package typex

import (
	"reflect"
)

func Deref(t Type) Type {
	for t.Kind() == reflect.Pointer && t.Name() == "" {
		t = t.Elem()
	}
	return t
}

func PosOfStructField(f StructField) int {
	if x, ok := f.(interface{ Pos() int }); ok {
		return x.Pos()
	}
	return 0
}
