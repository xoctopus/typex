package typex

import (
	"reflect"

	"github.com/xoctopus/typex/internal/x"
)

type (
	Type        = x.Type
	Method      = x.Method
	StructField = x.StructField
)

// Literal TODO handle literal of alias
// func Literal(t Type) string {
// 	switch tt := t.(type) {
// 	case *ttype:
// 		if tt.alias != nil {
// 			return "TODO"
// 		}
// 		return tt.TypeLit()
// 	case *rtype:
// 		// cannot detect alias of runtime type
// 		return ""
// 	default:
// 		if t == nil {
// 			return ""
// 		}
// 		panic("unexpected type: %T")
// 	}
// }

func Deref(t Type) Type {
	for t.Kind() == reflect.Pointer && t.Name() == "" {
		t = t.Elem()
	}
	return t
}
