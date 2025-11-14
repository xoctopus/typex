package testdata

import (
	"context"
	"os"

	"github.com/xoctopus/pkgx"
	"github.com/xoctopus/x/misc/must"
)

var Context context.Context

func init() {
	Context = context.Background()
	Context = pkgx.WithTests(Context)
	Context = pkgx.WithWorkdir(Context, must.NoErrorV(os.Getwd()))
}
