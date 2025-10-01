package pkgx_test

import (
	"go/types"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xoctopus/x/testx"

	. "github.com/xoctopus/typex/internal/pkgx"
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
		NewWithT(t).Expect(p.Name()).To(Equal(c.name))
		NewWithT(t).Expect(p.Path()).To(Equal(c.pkg))
		NewWithT(t).Expect(p.Unwrap().Path()).To(Equal(c.pkg))
		NewWithT(t).Expect(p.ID()).To(Equal(c.id))
	}
	NewWithT(t).Expect(New("")).To(BeNil())
	NewWithT(t).Expect(NewT(nil)).To(BeNil())
}

type (
	Named struct{}
	Alias = Named
)

func TestLookup(t *testing.T) {
	path := "github.com/xoctopus/typex/internal/pkgx_test"
	p := NewT(types.NewPackage(path, "pkgx_test"))

	named, exists := Lookup[*types.Named](p, "Undefined")
	NewWithT(t).Expect(exists).To(BeFalse())

	named, exists = Lookup[*types.Named](p, "Named")
	NewWithT(t).Expect(exists).To(BeTrue())
	NewWithT(t).Expect(named.Obj().Name()).To(Equal("Named"))

	t.Run("ObjectNotFound", func(t *testing.T) {
		testx.ExpectPanic[error](
			t,
			func() { MustLookup[*types.Named](p, "Undefined") },
			testx.ErrorContains("object `Undefined` not found"),
		)
	})

	_, exists = Lookup[*types.Named](p, "Alias")
	NewWithT(t).Expect(exists).To(BeFalse())

	t.Run("TypeUnmatched", func(t *testing.T) {
		testx.ExpectPanic[error](
			t,
			func() { MustLookup[*types.Named](p, "Alias") },
			testx.ErrorEqual("object `Alias` is not a *types.Named type"),
		)
	})

	alias := MustLookup[*types.Alias](p, "Alias")
	NewWithT(t).Expect(types.Identical(alias, named)).To(BeTrue())

	named, exists = LookupByPath[*types.Named](path, "Named")
	NewWithT(t).Expect(exists).To(BeTrue())
	alias = MustLookupByPath[*types.Alias](path, "Alias")
	NewWithT(t).Expect(types.Identical(types.Unalias(alias), named)).To(BeTrue())

}

func TestLoad(t *testing.T) {
	path := "github.com/xoctopus/typex/internal/pkgx"
	pkg := Load(path)
	NewWithT(t).Expect(pkg.Path()).To(Equal(path))

	path = "github.com/xoctopus/typex/internal/pkgx_test"
	pkg = Load(path)
	NewWithT(t).Expect(pkg.Path()).To(Equal(path))

	NewWithT(t).Expect(pkg.Scope().Lookup("TestLoad")).NotTo(BeNil())
}
