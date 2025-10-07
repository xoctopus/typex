package pkgx_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

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
		Expect(t, pkgx.Wrap(c[0]), Equal(c[1]))
	}
	pkgx.Clear()
	for _, c := range cases {
		Expect(t, pkgx.Unwrap(c[1]), Equal(c[0]))
	}

	pkgx.Clear()
	Expect(t, pkgx.Wrap("xwrap_net"), Equal("xwrap_net"))
	Expect(t, pkgx.Unwrap("xwrap_net"), Equal("net"))
	Expect(t, pkgx.Wrap("net"), Equal("net"))
}
