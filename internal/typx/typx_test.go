package typx_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/dumper"
	"github.com/xoctopus/typx/internal/typx"
)

func TestLitType(t *testing.T) {
	for _, c := range LitTypeCases {
		t.Run(c.name, func(t *testing.T) {
			rt := typx.NewLitType(c.rt)
			Expect(t, rt.String(), Equal(c.origin))
			Expect(t, rt.PkgPath(), Equal(c.PkgPath))
			Expect(t, rt.Name(), Equal(c.Name))
			Expect(t, rt.Dump(context.Background()), Equal(c.origin))
			Expect(t, rt.Dump(dumper.CtxWrapID.With(context.Background(), true)), Equal(c.wrapped))
			Expect(t, rt.Dump(dumper.CtxWrapID.With(context.Background(), false)), Equal(c.origin))

			tt := typx.NewLitType(c.tt)
			Expect(t, tt.String(), Equal(c.origin))
			Expect(t, tt.PkgPath(), Equal(c.PkgPath))
			Expect(t, tt.Name(), Equal(c.Name))
			Expect(t, tt.Dump(context.Background()), Equal(c.origin))
			Expect(t, tt.Dump(dumper.CtxWrapID.With(context.Background(), true)), Equal(c.wrapped))
			Expect(t, tt.Dump(dumper.CtxWrapID.With(context.Background(), false)), Equal(c.origin))
		})
	}
	t.Run("HitCache", func(t *testing.T) {
		for _, c := range LitTypeCases {
			t.Run(c.name, func(t *testing.T) {
				rt := typx.NewLitType(c.rt)
				tt := typx.NewLitType(c.tt)

				Expect(t, rt.String(), Equal(c.origin))
				Expect(t, rt.PkgPath(), Equal(c.PkgPath))
				Expect(t, rt.Name(), Equal(c.Name))

				Expect(t, tt.String(), Equal(c.origin))
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
	t.Run("WithPkgNamer", func(t *testing.T) {
		dUnnamedStruct = `struct { ` +
			`A string; ` +
			`B int; ` +
			`renamed.Map "json:\"esc''{}[]\\\"\""; ` +
			`renamed.TypedArray[net.Addr]; ` +
			`C struct { ` +
			`renamed.TypedArray[struct { renamed.TypedArray[fmt.Stringer] }] ` +
			`}; ` +
			`D interface { ` +
			`Close() error; ` +
			`Read([]uint8) (int, error); ` +
			`String() string; ` +
			`Write([]uint8) (int, error) ` +
			`} ` +
			`}`

		for _, c := range LitTypeCases {
			if c.name == "TypedArrayUnnamedStruct" {
				expect := "renamed.TypedArray[" + dUnnamedStruct + "]"

				rt := typx.NewLitType(c.rt)
				Expect(t, rt.Dump(dumper.CtxPkgNamer.With(context.Background(), &PkgNamer{})), Equal(expect))

				tt := typx.NewLitType(c.tt)
				Expect(t, tt.Dump(dumper.CtxPkgNamer.With(context.Background(), &PkgNamer{})), Equal(expect))

				break
			}
		}
	})
}

type PkgNamer struct{}

func (p PkgNamer) PackageName(path string) string {
	if path == "github.com/xoctopus/typx/testdata" {
		return "renamed"
	}
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		return path[idx:]
	}
	return path
}
