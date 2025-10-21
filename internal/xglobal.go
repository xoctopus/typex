package internal

import (
	"go/types"
	"reflect"

	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/syncx"
)

func init() {
	for _, t := range builtins {
		id := t.String()

		g.wraps.Store(id, id)
		g.wraps.Store(t.rtyp, id)
		g.wraps.Store(t.ttyp, id)
		if alias := t.Alias(); alias != "" {
			g.wraps.Store(alias, id)
		}

		g.literals.Store(id, t)
		g.literals.Store(t.rtyp, t)
		g.literals.Store(t.ttyp, t)

		g.ttypes.Store(id, t.ttyp)
		g.ttypes.Store(t.rtyp, t.ttyp)
	}
}

func Global() TypeGlobal { return g }

type TypeGlobal interface {
	Wrap(any) string
	Literalize(any) Literal
	TType(any) types.Type
}

type global struct {
	wraps    syncx.Map[any, string]
	literals syncx.Map[any, Literal]
	ttypes   syncx.Map[any, types.Type]
}

var g = &global{
	wraps:    syncx.NewSmap[any, string](),
	literals: syncx.NewSmap[any, Literal](),
	ttypes:   syncx.NewSmap[any, types.Type](),
}

func (g *global) Wrap(key any) string {
	switch key.(type) {
	case reflect.Type, types.Type:
		return g.wrap(key)
	default:
		if key == nil {
			return ""
		}
		panic(errors.Errorf("invalid wrap key type, it must be `reflect.Type` or `types.Type`, but got `%T`", key))
	}
}

func (g *global) wrap(key any) string {
	if v, matched := g.wraps.Load(key); matched {
		return v
	}

	id := ""
	switch k := key.(type) {
	case string:
		if id = wrap(k); id != k {
			g.wraps.Store(id, id)
		}
	case reflect.Type:
		id = wrapRT(k)
	default:
		t, ok := key.(types.Type)
		must.BeTrueF(ok, "expect string, reflect.Type or types.Type, but got `%T`", key)

		matched := false
		g.wraps.Range(func(a any, s string) bool {
			if x, ok := a.(types.Type); ok && types.Identical(x, t) {
				matched, id = true, s
				return false
			}
			return true
		})
		if matched {
			return id
		}
		id = wrapTT(key.(types.Type))
	}

	g.wraps.Store(key, id)
	return id
}

func (g *global) Literalize(key any) Literal {
	switch key.(type) {
	case reflect.Type, types.Type:
		return g.literalize(key)
	default:
		if key == nil {
			return nil
		}
		panic(errors.Errorf("invalid literalize key type, it must be `reflect.Type` or `types.Type`, but got `%T`", key))
	}
}

func (g *global) literalize(key any) Literal {
	if v, matched := g.literals.Load(key); matched {
		return v
	}

	var u Literal
	switch k := key.(type) {
	case string:
		u = literalize(k)
	case reflect.Type:
		u = literalizeRT(k)
	default:
		t, ok := key.(types.Type)
		must.BeTrueF(ok, "expect string, reflect.Type or types.Type, but got `%T`", key)

		matched := false
		g.literals.Range(func(a any, literal Literal) bool {
			if x, ok := a.(types.Type); ok && types.Identical(x, t) {
				matched, u = true, literal
				return false
			}
			return true
		})
		if matched {
			return u
		}
		u = literalizeTT(t)
	}

	g.literals.Store(key, u)
	_ = g.wrap(u.String())
	_ = g.wrap(key)

	return u
}

func (g *global) TType(key any) types.Type {
	switch k := key.(type) {
	case reflect.Type:
		return g.literalize(k).TType()
	case Literal:
		return k.TType()
	default:
		panic(errors.Errorf("invalid ttype key type, it must be `reflect.Type` or `Literal` but got %T", k))
	}
}
