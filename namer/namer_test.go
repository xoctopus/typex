package namer_test

import (
	"context"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex/namer"
)

type PackageNamer struct{}

func (PackageNamer) Package(path string) string {
	return path
}

func TestPackageNamer(t *testing.T) {
	demo := &PackageNamer{}
	ctx := namer.WithContext(context.Background(), demo)

	Expect(t, namer.WithContext(ctx, nil), Equal(ctx))

	n, ok := namer.FromContext(ctx)
	Expect(t, ok, Equal(true))
	Expect(t, n, Equal[namer.PackageNamer](demo))

	ExpectPanic[error](t, func() { namer.MustFromContext(context.Background()) })
}
