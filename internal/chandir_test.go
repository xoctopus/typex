package internal_test

import (
	"go/ast"
	"go/types"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex/internal"
)

func TestNewChanDir(t *testing.T) {
	for _, v := range []struct {
		tc types.ChanDir
		c  internal.ChanDir
	}{
		{types.SendOnly, internal.SendDir},
		{types.RecvOnly, internal.RecvDir},
		{types.SendRecv, internal.BothDir},
	} {
		dir := internal.NewChanDir(v.tc)
		Expect(t, dir, Equal(v.c))
		Expect(t, dir.TypesChanDir(), Equal(v.tc))
	}

	for _, v := range []struct {
		rc reflect.ChanDir
		c  internal.ChanDir
	}{
		{reflect.SendDir, internal.SendDir},
		{reflect.RecvDir, internal.RecvDir},
		{reflect.BothDir, internal.BothDir},
	} {
		dir := internal.NewChanDir(v.rc)
		Expect(t, dir, Equal(v.c))
		Expect(t, dir.ReflectChanDir(), Equal(v.rc))
	}

	for _, v := range []struct {
		ac ast.ChanDir
		c  internal.ChanDir
	}{
		{ast.SEND, internal.SendDir},
		{ast.RECV, internal.RecvDir},
		{ast.ChanDir(3), internal.BothDir},
	} {
		dir := internal.NewChanDir(v.ac)
		Expect(t, dir, Equal(v.c))
		Expect(t, dir.AstChanDir(), Equal(v.ac))
	}

	for _, v := range []struct {
		c internal.ChanDir
		s string
	}{
		{internal.SendDir, "chan<- "},
		{internal.RecvDir, "<-chan "},
		{internal.BothDir, "chan "},
	} {
		dir := internal.NewChanDir(v.c)
		Expect(t, dir, Equal(v.c))
		Expect(t, dir.String(), Equal(v.s))
	}

	t.Run("InvalidInput", func(t *testing.T) {
		ExpectPanic(
			t,
			func() { internal.NewChanDir(1) },
			ErrorContains("invalid dir type"),
		)
	})
}
