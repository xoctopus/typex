package typx

import (
	"testing"

	. "github.com/xoctopus/x/testx"
)

func TestEncodePathIdent(t *testing.T) {
	Expect(t, EncodePath(""), Equal(""))
	Expect(t, DecodePath(""), Equal(""))
	Expect(t, EncodePath("a"), Equal("a"))

	Expect(t, DecodePath("a"), Equal("a"))

	path := "path/to/package-x_y.z/v10"
	wrap := EncodePath(path)
	Expect(t, wrap, Equal("path_to_package_x_y_z_v10"))

	Expect(t, DecodePath("ident"), Equal("ident"))

	Expect(t, EncodePath(path), Equal(wrap))
	Expect(t, DecodePath(wrap), Equal(path))
}
