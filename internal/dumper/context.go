package dumper

import "github.com/xoctopus/x/contextx"

var (
	CtxWrapID   = contextx.NewT[bool]()
	CtxPkgNamer = contextx.NewT[PkgNamer]()
)

type PkgNamer interface {
	PackageName(string) string
}
