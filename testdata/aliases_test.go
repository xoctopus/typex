package testdata

import (
	"fmt"
	"go/types"

	"github.com/xoctopus/pkgx"
)

func Example_aliases() {
	for _, name := range []string{
		"AliasInt",
		"AliasUnion",
		"AliasSerialized",
		"AliasWithTArg",
	} {
		x := pkgx.MustLookup[*types.Alias](Context, "github.com/xoctopus/typx/testdata", name)
		fmt.Println(x)

		// x.
	}

	// Output:
	// github.com/xoctopus/typx/testdata.AliasInt
	// github.com/xoctopus/typx/testdata.AliasUnion
	// github.com/xoctopus/typx/testdata.AliasSerialized
	// github.com/xoctopus/typx/testdata.AliasWithTArg[X github.com/xoctopus/typx/testdata.CanBeSerialized]
}
