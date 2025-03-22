package pkgx_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/xoctopus/typex/internal/pkgx"
)

func TestWrapAndUnwrap(t *testing.T) {
	cases := [][2]string{
		{"net", "net"},
		{"fmt", "fmt"},
		{"encoding/json", "xwrap_encoding_s_json"},
		{"github.com/path/to/pkg.Type", "xwrap_github_d_com_s_path_s_to_s_pkg_d_Type"},
		{"github.com/path/to/pkg_test.Type", "xwrap_github_d_com_s_path_s_to_s_pkg_u_test_d_Type"},
	}

	pkgx.Clear()
	for _, c := range cases {
		NewWithT(t).Expect(pkgx.Wrap(c[0])).To(Equal(c[1]))
	}
	pkgx.Clear()
	for _, c := range cases {
		NewWithT(t).Expect(pkgx.Unwrap(c[1])).To(Equal(c[0]))
	}

	pkgx.Clear()
	NewWithT(t).Expect(pkgx.Wrap("xwrap_net")).To(Equal("xwrap_net"))
	NewWithT(t).Expect(pkgx.Unwrap("xwrap_net")).To(Equal("net"))
	NewWithT(t).Expect(pkgx.Wrap("net")).To(Equal("net"))
}
