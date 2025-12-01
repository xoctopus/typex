package typx

import (
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/stringsx"
	"github.com/xoctopus/x/syncx"
)

var (
	gPath2Wrap = syncx.NewXmap[string, string]()
	gWrap2Path = syncx.NewXmap[string, string]()
)

func EncodePath(p string) (w string) {
	if p == "" || stringsx.ValidIdentifier(p) {
		return p
	}

	if x, ok := gPath2Wrap.Load(p); ok {
		return x
	}

	defer func() {
		gPath2Wrap.Store(p, w)
		gWrap2Path.Store(w, p)
	}()

	r := []rune(p)
	for i, c := range r {
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' ||
			c >= 'a' && c <= 'z' || c == '_') {
			r[i] = '_'
		}
	}
	w = string(r)
	return
}

func DecodePath(w string) string {
	x, ok := gWrap2Path.Load(w)
	if !ok {
		must.BeTrueF(
			w == "" || stringsx.ValidIdentifier(w),
			"`%s` must be an identifier, if not be encoded.", w,
		)
		return w
	}
	return x
}
