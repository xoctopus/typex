package gtypex_test

import (
	"go/types"
	"net"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex/internal"
	"github.com/xoctopus/typex/internal/gtypex"
	"github.com/xoctopus/typex/internal/pkgx"
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

var path = "github.com/xoctopus/typex/internal/gtypex_test"

func TestUnderlying(t *testing.T) {
	t.Run("SameUnderlying", func(t *testing.T) {
		sig := pkgx.MustLookup[*types.Signature](pkgx.New("errors"), "New")
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
		instUnderlying := gtypex.Instantiate(
			pkgx.MustLookupByPath[*types.Named](path, "Generic").Underlying(),
			[]types.Type{types.Typ[types.Int]},
		)
		underlyingInst := gtypex.Underlying(
			internal.Global().TType(reflect.TypeFor[Generic[int]]()),
		)
		Expect(t, instUnderlying.String(), Equal(underlyingInst.String()))
	})
}
