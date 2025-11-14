package typex

import "reflect"

func Deref(t Type) Type {
	for t.Kind() == reflect.Pointer && t.Name() == "" {
		t = t.Elem()
	}
	return t
}
