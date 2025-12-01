package typx_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/typx"
)

func TestPackageLoad(t *testing.T) {
	pkg := typx.Load("github.com/xoctopus/typx/pkg/typx_test")
	Expect(t, pkg.Path(), Equal("github.com/xoctopus/typx/pkg/typx_test"))

	ExpectPanic[error](t, func() {
		pkg = typx.Load("github.com/xoctopus/typx/pkg/typex")
	})
}
