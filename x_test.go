package typex_test

import (
	"go/types"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xoctopus/x/reflectx"
	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex"
	"github.com/xoctopus/typex/internal/pkgx"
	"github.com/xoctopus/typex/testdata"
)

var pkg = pkgx.New("github.com/xoctopus/typex/testdata")

func init() {
	testdata.RegisterInstantiations(
		func(v any) typex.Type {
			return typex.NewRType(reflectx.MustAssertType[reflect.Type](v))
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
		NewWithT(t).Expect(types.Identical(tt, types.Typ[types.Int])).To(BeTrue())
	})
	t.Run("InvalidInput", func(t *testing.T) {
		t.Run("Union", func(t *testing.T) {
			defer func() {
				testx.AssertRecoverEqual(t, recover(), "invalid NewTType by types.Type for `*types.Union`")
			}()
			tt := pkgx.MustLookup[*types.Named](pkg, "Float").Underlying().(*types.Interface).EmbeddedType(0)
			typex.NewTType(tt)
		})
		t.Run("Tuple", func(t *testing.T) {
			defer func() {
				testx.AssertRecoverEqual(t, recover(), "invalid NewTType by types.Type for `*types.Tuple`")
			}()
			tt := pkgx.MustLookup[*types.Named](pkg, "Compare").Underlying().(*types.Signature).Results()
			typex.NewTType(tt)
		})
		t.Run("TypeParam", func(t *testing.T) {
			defer func() {
				testx.AssertRecoverEqual(t, recover(), "invalid NewTType by types.Type for `*types.TypeParam`")
			}()
			tt := pkgx.MustLookup[*types.Named](pkg, "BTreeNode").TypeParams().At(0)
			typex.NewTType(tt)
		})
		defer func() {
			testx.AssertRecoverEqual(t, recover(), "invalid NewTType type `int`")
		}()
		typex.NewTType(1)
	})
}
