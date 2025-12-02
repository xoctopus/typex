package typx_test

import (
	"go/types"
	"reflect"
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	typi "github.com/xoctopus/typx/internal/typx"
	"github.com/xoctopus/typx/pkg/typx"
	"github.com/xoctopus/typx/testdata"
)

var path = "github.com/xoctopus/typx/testdata"

func init() {
	testdata.RegisterInstantiations(
		func(v any) typx.Type {
			t, ok := v.(reflect.Type)
			must.BeTrue(ok)
			return typx.NewRType(t)
		},
		func(v any) typx.Type {
			t, ok := v.(types.Type)
			must.BeTrue(ok)
			return typx.NewTType(t)
		},
	)
}

func TestX(t *testing.T) {
	for _, c := range testdata.Cases {
		t.Run(c.Name(), c.Run)
	}
}

func TestNewTType(t *testing.T) {
	t.Run("ReflectType", func(t *testing.T) {
		tt := typx.NewTType(types.Typ[types.Int]).Unwrap().(types.Type)
		Expect(t, types.Identical(tt, types.Typ[types.Int]), BeTrue())
	})
	t.Run("InvalidInput", func(t *testing.T) {
		pkg := typi.Load(path)
		t.Run("Union", func(t *testing.T) {
			tt := typi.Lookup[*types.Named](pkg, "Float").Underlying().(*types.Interface).EmbeddedType(0)
			ExpectPanic[error](t, func() { typx.NewTType(tt) })
		})
		t.Run("Tuple", func(t *testing.T) {
			tt := typi.Lookup[*types.Named](pkg, "Compare").Underlying().(*types.Signature).Results()
			ExpectPanic[error](t, func() { typx.NewTType(tt) })
		})
		t.Run("TypeParam", func(t *testing.T) {
			tt := typi.Lookup[*types.Named](pkg, "BTreeNode").TypeParams().At(0)
			ExpectPanic[error](t, func() { typx.NewTType(tt) })
		})
	})
}
