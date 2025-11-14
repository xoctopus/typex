package gtypex_test

import (
	"context"
	"go/types"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/xoctopus/pkgx"
	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex/internal"
	"github.com/xoctopus/typex/internal/gtypex"
	"github.com/xoctopus/typex/testdata"
)

type Generic[T any] struct {
	Basic        int
	Alias        any
	Array        [3]int
	Slice        []T
	Chan         chan T
	Interface    interface{ Value() T }
	Struct       struct{ v T }
	Map          map[string]T
	Pointer      *T
	Named        testdata.Tagged
	GenericNamed testdata.PassTypeParam[T, net.Addr]
}

func TestUnderlying(t *testing.T) {
	t.Run("SameUnderlying", func(t *testing.T) {
		sig := pkgx.MustLookup[*types.Signature](context.Background(), "errors", "New")
		for _, v := range []types.Type{
			types.Typ[types.Int],                                // basic
			types.NewArray(types.Typ[types.Int], 1),             // array
			types.NewChan(types.SendRecv, types.Typ[types.Int]), // chan
			sig.Results().At(0).Type().Underlying(),             // interface
			types.NewPointer(types.Typ[types.Int]),              // map
			sig,                                                 // signature
			types.NewSlice(types.Typ[types.Int]),                // slice
		} {
			underlying := gtypex.Underlying(v)
			Expect(t, types.Identical(underlying, v), BeTrue())
		}
	})

	t.Run("Generic", func(t *testing.T) {
		path := "github.com/xoctopus/typex/internal/gtypex_test"
		ctx := context.Background()

		ctx = pkgx.WithTests(ctx)
		ctx = pkgx.WithWorkdir(ctx, must.NoErrorV(os.Getwd()))
		ctx = pkgx.WithLoadMode(ctx, pkgx.DefaultLoadMode)

		generic := pkgx.MustLookup[*types.Named](ctx, path, "Generic").Underlying()

		instUnderlying := gtypex.Instantiate(
			// pkgx.MustLookup[*types.Named](ctx, path, "Generic").Underlying(),
			generic,
			[]types.Type{types.Typ[types.Int]},
		)
		underlyingInst := gtypex.Underlying(
			internal.Global().TType(ctx, reflect.TypeFor[Generic[int]]()),
		)
		Expect(t, instUnderlying.String(), Equal(underlyingInst.String()))
	})
}
