package typx_test

import (
	"go/types"
	"net"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/typx"
	"github.com/xoctopus/typx/testdata"
)

func TestInstantiate(t *testing.T) {
	t.Run("NoTypeArgs", func(t *testing.T) {
		for _, v := range []types.Type{
			tInt,
			types.NewArray(tInt, 1),
			types.NewChan(types.SendRecv, tInt),
			tError.Underlying(),
			types.NewPointer(tInt),
			typx.Lookup[*types.Signature](stdErrors, "New"),
			types.NewSlice(tString),
		} {
			underlying := typx.Underlying(v)
			Expect(t, types.Identical(underlying, v), BeTrue())
		}
	})

	t.Run("Generic", func(t *testing.T) {
		pkg := typx.Load("github.com/xoctopus/typx/testdata")
		typ := typx.Lookup[types.Type](pkg, "Generics")

		instantiated := typx.Instantiate(typ.Underlying(), tInt).(*types.Struct)

		assertions := map[string]reflect.Type{
			"AliasInt":   reflect.TypeFor[testdata.AliasInt](),
			"Array":      reflect.TypeFor[[1]int](),
			"Basic":      reflect.TypeFor[int](),
			"Chan":       reflect.TypeFor[chan int](),
			"ChanW":      reflect.TypeFor[<-chan int](),
			"ChanR":      reflect.TypeFor[chan<- int](),
			"Map":        reflect.TypeFor[map[string]int](),
			"Pointer":    reflect.TypeFor[*int](),
			"Slice":      reflect.TypeFor[[]int](),
			"Struct":     reflect.TypeFor[struct{ FuncT func() int }](),
			"Interface":  reflect.TypeFor[interface{ Value() int }](),
			"TypedArray": reflect.TypeFor[testdata.TypedArray[int]](),
			"TypedSlice": reflect.TypeFor[testdata.TypedSlice[net.Addr]](),
			"NoTArg":     reflect.TypeFor[testdata.Int](),
		}

		for v := range instantiated.Fields() {
			t.Run(v.Name(), func(t *testing.T) {
				rt := assertions[v.Name()]
				tt := typx.NewTTByRT(rt)

				Expect(t, tt.String(), Equal(v.Type().String()))
			})
		}

		underlying := typx.Underlying(
			typx.NewTTByRT(reflect.TypeFor[testdata.Generics[int]]()),
		)
		// t.Log(instantiated.String())
		// t.Log(underlying.String())

		Expect(t, instantiated.String(), Equal(underlying.String()))
	})
}
