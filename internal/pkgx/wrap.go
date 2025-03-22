package pkgx

import (
	"strings"

	"github.com/xoctopus/x/mapx"
)

const (
	prefix     = "xwrap_"
	underscore = "_u_"
	dot        = "_d_"
	slash      = "_s_"
)

var (
	p2w = mapx.NewSmap[string, string]()
	w2p = mapx.NewSmap[string, string]()
)

func Clear() {
	p2w.Clear()
	w2p.Clear()
}

func Unwrap(w string) string {
	if v, ok := w2p.Load(w); ok {
		return v
	}
	if _, ok := p2w.Load(w); ok {
		return w
	}

	if strings.Contains(w, ".") || strings.Contains(w, "/") {
		return w
	}

	p := w
	p = strings.TrimPrefix(p, prefix)
	p = strings.ReplaceAll(p, slash, "/")
	p = strings.ReplaceAll(p, dot, ".")
	p = strings.ReplaceAll(p, underscore, "_")

	if !strings.Contains(p, ".") && !strings.Contains(p, "/") {
		w = p
	}

	p2w.Store(p, w)
	w2p.Store(w, p)

	return p
}

func Wrap(p string) string {
	if w, ok := p2w.Load(p); ok {
		return w
	}

	if strings.HasPrefix(p, prefix) {
		return p
	}

	if !strings.Contains(p, ".") && !strings.Contains(p, "/") {
		return p
	}

	w := p
	w = strings.ReplaceAll(w, "_", underscore)
	w = strings.ReplaceAll(w, ".", dot)
	w = strings.ReplaceAll(w, "/", slash)
	w = prefix + w

	p2w.Store(p, w)
	w2p.Store(w, p)

	return w
}
