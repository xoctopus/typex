package pkgutil_test

import (
	"go/types"
	"testing"

	. "github.com/xoctopus/x/testx"

	. "github.com/xoctopus/typex/pkgutil"
)

func TestNewPackage(t *testing.T) {
	cases := []struct {
		path string
		pkg  string
		name string
		id   string
	}{
		{
			path: "github.com/xoctopus/x/ptrx",
			pkg:  "github.com/xoctopus/x/ptrx",
			name: "ptrx",
			id:   "xwrap_github_d_com_s_xoctopus_s_x_s_ptrx",
		},
		{
			path: "github.com/xoctopus/x/ptrx",
			pkg:  "github.com/xoctopus/x/ptrx",
			name: "ptrx",
			id:   "xwrap_github_d_com_s_xoctopus_s_x_s_ptrx",
		},
		{
			path: "xwrap_github_d_com_s_xoctopus_s_x_s_ptrx",
			pkg:  "github.com/xoctopus/x/ptrx",
			name: "ptrx",
			id:   "xwrap_github_d_com_s_xoctopus_s_x_s_ptrx",
		},
		{
			path: "encoding/json",
			pkg:  "encoding/json",
			name: "json",
			id:   "xwrap_encoding_s_json",
		},
		{
			path: "net",
			pkg:  "net",
			name: "net",
			id:   "net",
		},
		{
			path: "xwrap_net",
			pkg:  "net",
			name: "net",
			id:   "net",
		},
	}

	for _, c := range cases {
		p := New(c.path)
		Expect(t, p.Name(), Equal(c.name))
		Expect(t, p.Path(), Equal(c.pkg))
		Expect(t, p.Unwrap().Path(), Equal(c.pkg))
		Expect(t, p.ID(), Equal(c.id))
	}
	Expect(t, New(""), BeNil[Package]())
	Expect(t, NewT(nil), BeNil[Package]())
}

type (
	Named struct{}
	Alias = Named
)

func TestLookup(t *testing.T) {
	path := "github.com/xoctopus/typex/pkgutil_test"
	p := NewT(types.NewPackage(path, "pkgutil_test"))

	_, exists := Lookup[*types.Named](p, "Undefined")
	Expect(t, exists, BeFalse())

	named, exists := Lookup[*types.Named](p, "Named")
	Expect(t, exists, BeTrue())
	Expect(t, named.Obj().Name(), Equal("Named"))

	t.Run("ObjectNotFound", func(t *testing.T) {
		ExpectPanic[error](
			t,
			func() { MustLookup[*types.Named](p, "Undefined") },
			ErrorContains("object `Undefined` not found"),
		)
	})

	_, exists = Lookup[*types.Named](p, "Alias")
	Expect(t, exists, BeFalse())

	t.Run("TypeUnmatched", func(t *testing.T) {
		ExpectPanic[error](
			t,
			func() { MustLookup[*types.Named](p, "Alias") },
			ErrorEqual("object `Alias` is not a *types.Named type"),
		)
	})

	alias := MustLookup[*types.Alias](p, "Alias")
	Expect(t, types.Identical(alias, named), BeTrue())

	named, exists = LookupByPath[*types.Named](path, "Named")
	Expect(t, exists, BeTrue())
	alias = MustLookupByPath[*types.Alias](path, "Alias")
	Expect(t, types.Identical(types.Unalias(alias), named), BeTrue())

}

func TestLoad(t *testing.T) {
	path := "github.com/xoctopus/typex/pkgutil"
	pkg := New(path)
	Expect(t, pkg.Path(), Equal(path))

	path = "github.com/xoctopus/typex/pkgutil_test"
	pkg = New(path)
	Expect(t, pkg.Path(), Equal(path))

	Expect(t, pkg.Scope().Lookup("TestLoad"), NotBeNil[types.Object]())
}
