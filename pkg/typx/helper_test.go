package typx_test

import (
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	lit "github.com/xoctopus/typex/internal/typx"
	"github.com/xoctopus/typex/pkg/typx"
)

type (
	T1 struct {
		A string
	}
	T2 *int
	T3 *T1
)

func TestDeref(t *testing.T) {
	rt := typx.NewRType(reflect.TypeOf(T1{}))
	dt := typx.Deref(rt)
	Expect(t, dt.String(), Equal(rt.String()))

	rt = typx.NewRType(reflect.TypeOf(*new(T2)))
	dt = typx.Deref(rt)
	Expect(t, dt.String(), Equal(rt.String()))

	rt = typx.NewRType(reflect.TypeOf(new(int)))
	dt = typx.Deref(rt)
	Expect(t, dt.String(), Equal("int"))

	rt = typx.NewRType(reflect.TypeFor[*****int]())
	dt = typx.Deref(rt)
	Expect(t, dt.String(), Equal("int"))

	rt = typx.NewRType(reflect.TypeFor[T3]())
	dt = typx.Deref(rt)
	Expect(t, dt.String(), Equal("github.com/xoctopus/typex/pkg/typx_test.T3"))
}

func TestPosOfStructField(t *testing.T) {
	tt := typx.NewTType(lit.NewLitType(reflect.TypeOf(T1{})).Type())
	Expect(t, typx.PosOfStructField(tt.Field(0)), NotEqual(0))

	rt := typx.NewRType(reflect.TypeOf(T1{}))
	Expect(t, typx.PosOfStructField(rt.Field(0)), Equal(0))
}
