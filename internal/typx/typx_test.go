package typx_test

import (
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/typx"
)

func TestLitType(t *testing.T) {
	for _, c := range LitTypeCases {
		t.Run(c.name, func(t *testing.T) {
			rt := typx.NewLitType(c.rt)
			tt := typx.NewLitType(c.tt)

			Expect(t, rt.String(), Equal(c.origin))
			Expect(t, rt.Literal(), Equal(c.expect))
			Expect(t, rt.PkgPath(), Equal(c.PkgPath))
			Expect(t, rt.Name(), Equal(c.Name))

			Expect(t, tt.String(), Equal(c.origin))
			Expect(t, tt.Literal(), Equal(c.expect))
			Expect(t, tt.PkgPath(), Equal(c.PkgPath))
			Expect(t, tt.Name(), Equal(c.Name))
		})
	}
	t.Run("HitCache", func(t *testing.T) {
		for _, c := range LitTypeCases {
			t.Run(c.name, func(t *testing.T) {
				rt := typx.NewLitType(c.rt)
				tt := typx.NewLitType(c.tt)

				Expect(t, rt.String(), Equal(c.origin))
				Expect(t, rt.Literal(), Equal(c.expect))
				Expect(t, rt.PkgPath(), Equal(c.PkgPath))
				Expect(t, rt.Name(), Equal(c.Name))

				Expect(t, tt.String(), Equal(c.origin))
				Expect(t, tt.Literal(), Equal(c.expect))
				Expect(t, tt.PkgPath(), Equal(c.PkgPath))
				Expect(t, tt.Name(), Equal(c.Name))

				if c.Name == "int" {
					Expect(t, rt.Kind(), Equal(reflect.Int))
					Expect(t, tt.Kind(), Equal(reflect.Int))
				}
			})
		}
		ExpectPanic[error](t, func() { typx.NewLitType(nil) })
	})
}
