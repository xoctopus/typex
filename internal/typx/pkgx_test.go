package typx_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex/internal/typx"
)

func TestPackageLoad(t *testing.T) {
	pkg := typx.Load("github.com/xoctopus/typex/pkg/typx_test")
	Expect(t, pkg.Path(), Equal("github.com/xoctopus/typex/pkg/typx_test"))

	ExpectPanic[error](t, func() {
		pkg = typx.Load("github.com/xoctopus/typex/pkg/typex")
	})
}
