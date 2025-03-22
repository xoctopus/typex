package internal_test

import (
	"go/ast"
	"go/types"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex/internal"
)

func TestNewChanDir(t *testing.T) {
	for _, v := range [][2]any{
		{types.SendOnly, internal.SendDir},
		{types.RecvOnly, internal.RecvDir},
		{types.SendRecv, internal.BothDir},
	} {
		dir := internal.NewChanDir(v[0])
		NewWithT(t).Expect(dir).To(Equal(v[1]))
		NewWithT(t).Expect(dir.TypesChanDir()).To(Equal(v[0]))
	}

	for _, v := range [][2]any{
		{reflect.SendDir, internal.SendDir},
		{reflect.RecvDir, internal.RecvDir},
		{reflect.BothDir, internal.BothDir},
	} {
		dir := internal.NewChanDir(v[0])
		NewWithT(t).Expect(dir).To(Equal(v[1]))
		NewWithT(t).Expect(dir.ReflectChanDir()).To(Equal(v[0]))
	}

	for _, v := range [][2]any{
		{ast.SEND, internal.SendDir},
		{ast.RECV, internal.RecvDir},
		{ast.ChanDir(3), internal.BothDir},
	} {
		dir := internal.NewChanDir(v[0])
		NewWithT(t).Expect(dir).To(Equal(v[1]))
		NewWithT(t).Expect(dir.AstChanDir()).To(Equal(v[0]))
	}

	for _, v := range [][2]any{
		{internal.SendDir, "chan<- "},
		{internal.RecvDir, "<-chan "},
		{internal.BothDir, "chan "},
	} {
		dir := internal.NewChanDir(v[0])
		NewWithT(t).Expect(dir).To(Equal(v[0]))
		NewWithT(t).Expect(dir.String()).To(Equal(v[1]))
	}

	t.Run("InvalidInput", func(t *testing.T) {
		defer func() {
			testx.AssertRecoverContains(t, recover(), "invalid dir type")
		}()
		internal.NewChanDir(1)
	})
}
