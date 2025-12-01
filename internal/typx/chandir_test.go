package typx_test

import (
	"go/ast"
	"go/types"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex/internal/typx"
)

func TestChanDir(t *testing.T) {
	ExpectPanic[error](t, func() { typx.ChanDir(reflect.ChanDir(4)) })

	Expect(t, typx.ChanDir(reflect.SendDir), Equal("chan<- "))
	Expect(t, typx.ChanDir(reflect.RecvDir), Equal("<-chan "))
	Expect(t, typx.ChanDir(reflect.BothDir), Equal("chan "))

	Expect(t, typx.ChanDir(types.SendOnly), Equal("chan<- "))
	Expect(t, typx.ChanDir(types.RecvOnly), Equal("<-chan "))
	Expect(t, typx.ChanDir(types.SendRecv), Equal("chan "))

	Expect(t, typx.ChanDir(ast.SEND), Equal("chan<- "))
	Expect(t, typx.ChanDir(ast.RECV), Equal("<-chan "))
	Expect(t, typx.ChanDir(ast.SEND|ast.RECV), Equal("chan "))
}
