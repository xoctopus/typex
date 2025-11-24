package testdata

import (
	"context"
	"os"

	"github.com/xoctopus/pkgx"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"
)

var Context context.Context

func init() {
	Context = contextx.Compose(
		pkgx.CtxLoadTests.Carry(true),
		pkgx.CtxWorkdir.Carry(must.NoErrorV(os.Getwd())),
	)(context.Background())
}
