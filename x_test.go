package typex_test

import (
	"context"
	"go/types"
	"reflect"
	"testing"

	"github.com/xoctopus/pkgx"
	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex"
	"github.com/xoctopus/typex/testdata"
)

var (
	ctx  = testdata.Context
	path = "github.com/xoctopus/typex/testdata"
)

func init() {
	testdata.RegisterInstantiations(
		func(ctx context.Context, v any) typex.Type {
			t, ok := v.(reflect.Type)
			must.BeTrue(ok)
			return typex.NewRType(ctx, t)
		},
		typex.NewTType,
	)
}

func TestX(t *testing.T) {
	for _, c := range testdata.Cases {
		t.Run(c.Name(), c.Run)
	}
}

func TestNewTType(t *testing.T) {
	t.Run("ReflectType", func(t *testing.T) {
		tt := typex.NewTType(ctx, reflect.TypeFor[int]()).Unwrap().(types.Type)
		Expect(t, types.Identical(tt, types.Typ[types.Int]), BeTrue())
	})
	t.Run("InvalidInput", func(t *testing.T) {
		t.Run("Union", func(t *testing.T) {
			tt := pkgx.MustLookup[*types.Named](ctx, path, "Float").Underlying().(*types.Interface).EmbeddedType(0)
			ExpectPanic[error](
				t,
				func() { typex.NewTType(ctx, tt) },
				ErrorEqual("invalid NewTType by types.Type for `*types.Union`"),
			)
		})
		t.Run("Tuple", func(t *testing.T) {
			tt := pkgx.MustLookup[*types.Named](ctx, path, "Compare").Underlying().(*types.Signature).Results()
			ExpectPanic[error](
				t,
				func() { typex.NewTType(ctx, tt) },
				ErrorEqual("invalid NewTType by types.Type for `*types.Tuple`"),
			)
		})
		t.Run("TypeParam", func(t *testing.T) {
			tt := pkgx.MustLookup[*types.Named](ctx, path, "BTreeNode").TypeParams().At(0)
			ExpectPanic[error](
				t,
				func() { typex.NewTType(ctx, tt) },
				ErrorEqual("invalid NewTType by types.Type for `*types.TypeParam`"),
			)
		})

		ExpectPanic[error](
			t,
			func() { typex.NewTType(ctx, 1) },
			ErrorEqual("invalid NewTType type `int`"),
		)
	})
}
