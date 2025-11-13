package typex_test

import (
	"go/types"
	"reflect"
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex"
	"github.com/xoctopus/typex/pkgutil"
	"github.com/xoctopus/typex/testdata"
)

var pkg = pkgutil.New("github.com/xoctopus/typex/testdata")

func init() {
	testdata.RegisterInstantiations(
		func(v any) typex.Type {
			t, ok := v.(reflect.Type)
			must.BeTrue(ok)
			return typex.NewRType(t)
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
		tt := typex.NewTType(reflect.TypeFor[int]()).Unwrap().(types.Type)
		Expect(t, types.Identical(tt, types.Typ[types.Int]), BeTrue())
	})
	t.Run("InvalidInput", func(t *testing.T) {
		t.Run("Union", func(t *testing.T) {
			tt := pkgutil.MustLookup[*types.Named](pkg, "Float").Underlying().(*types.Interface).EmbeddedType(0)
			ExpectPanic[error](
				t,
				func() { typex.NewTType(tt) },
				ErrorEqual("invalid NewTType by types.Type for `*types.Union`"),
			)
		})
		t.Run("Tuple", func(t *testing.T) {
			tt := pkgutil.MustLookup[*types.Named](pkg, "Compare").Underlying().(*types.Signature).Results()
			ExpectPanic[error](
				t,
				func() { typex.NewTType(tt) },
				ErrorEqual("invalid NewTType by types.Type for `*types.Tuple`"),
			)
		})
		t.Run("TypeParam", func(t *testing.T) {
			tt := pkgutil.MustLookup[*types.Named](pkg, "BTreeNode").TypeParams().At(0)
			ExpectPanic[error](
				t,
				func() { typex.NewTType(tt) },
				ErrorEqual("invalid NewTType by types.Type for `*types.TypeParam`"),
			)
		})

		ExpectPanic[error](
			t,
			func() { typex.NewTType(1) },
			ErrorEqual("invalid NewTType type `int`"),
		)
	})
}
