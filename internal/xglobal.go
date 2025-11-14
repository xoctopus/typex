package internal

import (
	"context"
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
	Wrap(context.Context, any) string
	Literalize(context.Context, any) Literal
	TType(context.Context, any) types.Type
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

func (g *global) Wrap(ctx context.Context, key any) string {
	switch key.(type) {
	case reflect.Type, types.Type:
		return g.wrap(ctx, key)
	default:
		if key == nil {
			return ""
		}
		panic(errors.Errorf("invalid wrap key type, it must be `reflect.Type` or `types.Type`, but got `%T`", key))
	}
}

func (g *global) wrap(ctx context.Context, key any) string {
	if v, matched := g.wraps.Load(key); matched {
		return v
	}

	id := ""
	switch k := key.(type) {
	case string:
		if id = wrap(ctx, k); id != k {
			g.wraps.Store(id, id)
		}
	case reflect.Type:
		id = wrapRT(ctx, k)
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
		id = wrapTT(ctx, key.(types.Type))
	}

	g.wraps.Store(key, id)
	return id
}

func (g *global) Literalize(ctx context.Context, key any) Literal {
	switch key.(type) {
	case reflect.Type, types.Type:
		return g.literalize(ctx, key)
	default:
		if key == nil {
			return nil
		}
		panic(errors.Errorf("invalid literalize key type, it must be `reflect.Type` or `types.Type`, but got `%T`", key))
	}
}

func (g *global) literalize(ctx context.Context, key any) Literal {
	if v, matched := g.literals.Load(key); matched {
		return v
	}

	var u Literal
	switch k := key.(type) {
	case string:
		u = literalize(ctx, k)
	case reflect.Type:
		u = literalizeRT(ctx, k)
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
		u = literalizeTT(ctx, t)
	}

	g.literals.Store(key, u)
	_ = g.wrap(ctx, u.String())
	_ = g.wrap(ctx, key)

	return u
}

func (g *global) TType(ctx context.Context, key any) types.Type {
	switch k := key.(type) {
	case reflect.Type:
		return g.literalize(ctx, k).TType(ctx)
	case Literal:
		return k.TType(ctx)
	default:
		panic(errors.Errorf("invalid ttype key type, it must be `reflect.Type` or `Literal` but got %T", k))
	}
}
