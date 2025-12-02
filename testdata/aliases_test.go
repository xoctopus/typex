package testdata

import (
	"fmt"
	"go/types"

	typi "github.com/xoctopus/typx/internal/typx"
)

func Example_aliases() {
	for _, name := range []string{
		"AliasInt",
		"AliasUnion",
		"AliasSerialized",
		"AliasWithTArg",
	} {
		x := typi.Lookup[*types.Alias](typi.Load("github.com/xoctopus/typx/testdata"), name)
		fmt.Println(x)

		// x.
	}

	// Output:
	// github.com/xoctopus/typx/testdata.AliasInt
	// github.com/xoctopus/typx/testdata.AliasUnion
	// github.com/xoctopus/typx/testdata.AliasSerialized
	// github.com/xoctopus/typx/testdata.AliasWithTArg[X github.com/xoctopus/typx/testdata.CanBeSerialized]
}
