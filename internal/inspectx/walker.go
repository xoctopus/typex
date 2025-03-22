package inspectx

import (
	"go/types"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"
)

type Walker []types.Type

func (w *Walker) IsVisited(t types.Type) bool {
	must.BeTrue(!reflectx.CanCast[*types.Alias](t))
	for _, k := range *w {
		if types.Identical(k, t) {
			return true
		}
	}
	return false
}

func (w *Walker) Visited(t types.Type) {
	must.BeTrue(!reflectx.CanCast[*types.Alias](t))
	if n := len(*w); n > 0 {
		must.BeTrue(types.Identical((*w)[n-1], t))
		*w = (*w)[: n-1 : n]
	}
}

func (w *Walker) Visit(t types.Type) {
	must.BeTrue(!reflectx.CanCast[*types.Alias](t))
	*w = append(*w, t)
}
