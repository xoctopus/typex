package typx

import (
	"go/ast"
	"go/types"
	"reflect"

	"github.com/xoctopus/x/misc/must"
)

var (
	dirs = map[any]string{
		reflect.RecvDir:     "<-chan ",
		reflect.SendDir:     "chan<- ",
		reflect.BothDir:     "chan ",
		types.RecvOnly:      "<-chan ",
		types.SendOnly:      "chan<- ",
		types.SendRecv:      "chan ",
		ast.RECV:            "<-chan ",
		ast.SEND:            "chan<- ",
		ast.RECV | ast.SEND: "chan ",
	}
	tdirs = map[string]types.ChanDir{
		"<-chan ": types.RecvOnly,
		"chan<- ": types.SendOnly,
		"chan ":   types.SendRecv,
	}
)

func ChanDir(c any) string {
	s, ok := dirs[c]
	must.BeTrue(ok)
	return s
}

func TChanDir(c any) types.ChanDir {
	return tdirs[ChanDir(c)]
}
