package namer

import (
	"context"

	"github.com/xoctopus/x/contextx"
)

type PackageNamer interface {
	Package(id string) (name string)
}

// type Namers map[string]PackageNamer

var ctx = contextx.NewT[PackageNamer]()

func FromContext(parent context.Context) (PackageNamer, bool) {
	return ctx.From(parent)
}

func MustFromContext(parent context.Context) PackageNamer {
	return ctx.MustFrom(parent)
}

func WithContext(child context.Context, namer PackageNamer) context.Context {
	if _, ok := FromContext(child); !ok {
		return ctx.With(child, namer)
	}
	return child
}
