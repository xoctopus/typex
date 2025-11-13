package typex_test

import (
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex"
)

type (
	T1 struct{}
	T2 *int
	T3 *T1
)

func TestDeref(t *testing.T) {
	rt := typex.NewRType(reflect.TypeOf(T1{}))
	dt := typex.Deref(rt)
	Expect(t, dt.String(), Equal(rt.String()))

	rt = typex.NewRType(reflect.TypeOf(*new(T2)))
	dt = typex.Deref(rt)
	Expect(t, dt.String(), Equal(rt.String()))

	rt = typex.NewRType(reflect.TypeOf(new(int)))
	dt = typex.Deref(rt)
	Expect(t, dt.String(), Equal("int"))

	rt = typex.NewRType(reflect.TypeFor[*****int]())
	dt = typex.Deref(rt)
	Expect(t, dt.String(), Equal("int"))

	rt = typex.NewRType(reflect.TypeFor[T3]())
	dt = typex.Deref(rt)
	Expect(t, dt.String(), Equal("github.com/xoctopus/typex_test.T3"))
}
