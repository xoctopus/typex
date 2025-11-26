package typex_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/xoctopus/pkgx"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex"
)

type (
	T1 struct {
		A string
	}
	T2 *int
	T3 *T1
)

func TestDeref(t *testing.T) {
	ctx := contextx.Compose(
		pkgx.CtxLoadTests.Carry(true),
		pkgx.CtxWorkdir.Carry(must.NoErrorV(os.Getwd())),
	)(context.Background())

	rt := typex.NewRType(ctx, reflect.TypeOf(T1{}))
	dt := typex.Deref(rt)
	Expect(t, dt.String(), Equal(rt.String()))

	rt = typex.NewRType(ctx, reflect.TypeOf(*new(T2)))
	dt = typex.Deref(rt)
	Expect(t, dt.String(), Equal(rt.String()))

	rt = typex.NewRType(ctx, reflect.TypeOf(new(int)))
	dt = typex.Deref(rt)
	Expect(t, dt.String(), Equal("int"))

	rt = typex.NewRType(ctx, reflect.TypeFor[*****int]())
	dt = typex.Deref(rt)
	Expect(t, dt.String(), Equal("int"))

	rt = typex.NewRType(ctx, reflect.TypeFor[T3]())
	dt = typex.Deref(rt)
	Expect(t, dt.String(), Equal("github.com/xoctopus/typex_test.T3"))
}

func TestPosOfStructField(t *testing.T) {
	ctx := contextx.Compose(
		pkgx.CtxLoadTests.Carry(true),
		pkgx.CtxWorkdir.Carry(must.NoErrorV(os.Getwd())),
	)(context.Background())

	tt := typex.NewTType(ctx, reflect.TypeOf(T1{}))
	Expect(t, typex.PosOfStructField(tt.Field(0)), NotEqual(0))

	rt := typex.NewRType(ctx, reflect.TypeOf(T1{}))
	Expect(t, typex.PosOfStructField(rt.Field(0)), Equal(0))
}
