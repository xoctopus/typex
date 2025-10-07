package testdata

import (
	"fmt"
	"go/types"

	"github.com/xoctopus/typex/internal/pkgx"
)

func Example_aliases() {
	pkg := pkgx.New("github.com/xoctopus/typex/testdata")

	for _, name := range []string{
		"AliasInt",
		"AliasUnion",
		"AliasSerialized",
		"AliasWithTArg",
	} {
		x := pkgx.MustLookup[*types.Alias](pkg, name)
		fmt.Println(x)

		// x.
	}

	// Output:
	// github.com/xoctopus/typex/testdata.AliasInt
	// github.com/xoctopus/typex/testdata.AliasUnion
	// github.com/xoctopus/typex/testdata.AliasSerialized
	// github.com/xoctopus/typex/testdata.AliasWithTArg[X github.com/xoctopus/typex/testdata.CanBeSerialized]
}
