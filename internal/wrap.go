package internal

import (
	"fmt"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/typex/internal/parsex"
	"github.com/xoctopus/typex/pkgutil"
)

var (
	bracketed = parsex.Bracketed
	separate  = parsex.Separate
	fieldInfo = parsex.FieldInfo
)

func wrap(id string) (wrapped string) {
	// slice: []elem / array: [len]elem
	if strings.HasPrefix(id, "[") {
		idx, _, r := bracketed(id, '[')
		ele := id[r+1:]
		return "[" + idx + "]" + g.wrap(ele)
	}

	// map: map[key]elem
	if strings.HasPrefix(id, "map[") {
		key, _, r := bracketed(id, '[')
		ele := id[r+1:]
		return "map[" + g.wrap(key) + "]" + g.wrap(ele)
	}

	// chan elem
	if strings.HasPrefix(id, "chan ") {
		return "chan " + g.wrap(id[5:])
	}

	// chan<- elem
	if strings.HasPrefix(id, "chan<- ") {
		return "chan<- " + g.wrap(id[7:])
	}

	// <-chan elem
	if strings.HasPrefix(id, "<-chan ") {
		return "<-chan " + g.wrap(id[7:])
	}

	// struct: struct { fields... }
	if strings.HasPrefix(id, "struct {") {
		fields, _, _ := bracketed(id, '{')
		if len(fields) == 0 {
			return "struct {}"
		}

		b := strings.Builder{}
		b.WriteString("struct { ")
		for i, f := range separate(fields, ';') {
			if i > 0 {
				b.WriteString("; ")
			}
			name, typ, tag := fieldInfo(f)
			if len(name) > 0 {
				b.WriteString(name)
				b.WriteString(" ")
			}
			b.WriteString(g.wrap(typ))
			if len(tag) > 0 {
				b.WriteString(" ")
				b.WriteString(strconv.Quote(tag))
			}
		}
		b.WriteString(" }")
		return b.String()
	}

	// interface: interface { methods... }
	if strings.HasPrefix(id, "interface {") {
		methods, _, _ := bracketed(id, '{')
		must.BeTrue(len(methods) > 0)

		b := strings.Builder{}
		b.WriteString("interface { ")
		for i, m := range separate(methods, ';') {
			if i > 0 {
				b.WriteString("; ")
			}
			idx := strings.Index(m, "(")
			name := m[0:idx]
			typ := g.wrap("func" + m[idx:])
			b.WriteString(name + typ[4:])
		}
		b.WriteString(" }")
		return b.String()
	}

	// func: func(params...) results
	if strings.HasPrefix(id, "func(") {
		b := strings.Builder{}
		b.WriteString("func(")

		params, _, pr := bracketed(id, '(')
		for i, p := range separate(params, ',') {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(g.wrap(p))
		}
		b.WriteString(")")

		if pr == len(id)-1 {
			return b.String()
		}

		b.WriteString(" ")
		id = strings.TrimSpace(id[pr+1:])
		if id[0] != '(' {
			b.WriteString(id)
			return b.String()
		}

		b.WriteString("(")
		results, _, _ := bracketed(id, '(')
		for i, r := range separate(results, ',') {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(g.wrap(r))
		}
		b.WriteString(")")
		return b.String()
	}

	// pointer: *elem
	if strings.HasPrefix(id, "*") {
		return "*" + g.wrap(id[1:])
	}

	// variadic: ...elem
	if strings.HasPrefix(id, "...") {
		return "..." + g.wrap(id[3:])
	}

	// named: package_path.typename[type arguments...]
	path, name, targs := "", "", ""
	if v, l, r := bracketed(id, '['); l > 0 && r > 0 {
		targs = v
		id = id[0:l]
	}
	dot := strings.LastIndex(id, ".")
	must.BeTrue(dot != -1)
	path, name = id[0:dot], id[dot+1:]
	b := strings.Builder{}
	b.WriteString(pkgutil.New(path).ID())
	b.WriteString(".")
	b.WriteString(name)
	if len(targs) > 0 {
		b.WriteString("[")
		for i, targ := range separate(targs, ',') {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(g.wrap(targ))
		}
		b.WriteString("]")
	}
	return b.String()
}

func wrapRT(t reflect.Type) string {
	if t.Name() != "" {
		return g.wrap(t.PkgPath() + "." + t.Name())
	}

	switch t.Kind() {
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), g.wrap(t.Elem()))
	case reflect.Chan:
		return fmt.Sprintf("%s%s", NewChanDir(t.ChanDir()), g.wrap(t.Elem()))
	case reflect.Func:
		b := strings.Builder{}
		b.WriteString("func(")
		for i := range t.NumIn() {
			if i > 0 {
				b.WriteString(", ")
			}
			if i == t.NumIn()-1 && t.IsVariadic() {
				b.WriteString("...")
				b.WriteString(g.wrap(t.In(i).Elem()))
				break
			}
			b.WriteString(g.wrap(t.In(i)))
		}
		b.WriteString(")")
		if t.NumOut() == 0 {
			return b.String()
		}
		b.WriteString(" ")
		if t.NumOut() == 1 {
			b.WriteString(g.wrap(t.Out(0)))
			return b.String()
		}
		b.WriteString("(")
		for i := range t.NumOut() {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(g.wrap(t.Out(i)))
		}
		b.WriteString(")")
		return b.String()
	case reflect.Interface:
		b := strings.Builder{}
		b.WriteString("interface { ")
		for i := range t.NumMethod() {
			if i > 0 {
				b.WriteString("; ")
			}
			m := t.Method(i)
			f := g.wrap(m.Type)
			b.WriteString(m.Name)
			b.WriteString(f[4:])
		}
		b.WriteString(" }")
		return b.String()
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", g.wrap(t.Key()), g.wrap(t.Elem()))
	case reflect.Pointer:
		return "*" + g.wrap(t.Elem())
	case reflect.Slice:
		return "[]" + g.wrap(t.Elem())
	default:
		must.BeTrueF(t.Kind() == reflect.Struct, "unexpected kind %s", t.Kind())
		if t.NumField() == 0 {
			return "struct {}"
		}
		b := strings.Builder{}
		b.WriteString("struct { ")
		for i := range t.NumField() {
			if i > 0 {
				b.WriteString("; ")
			}
			f := t.Field(i)
			if !f.Anonymous {
				b.WriteString(f.Name)
				b.WriteString(" ")
			}
			b.WriteString(g.wrap(f.Type))
			if len(f.Tag) > 0 {
				b.WriteString(" ")
				b.WriteString(strconv.Quote(string(f.Tag)))
			}
		}
		b.WriteString(" }")
		return b.String()
	}
}

func wrapTT(t types.Type) string {
	switch x := t.(type) {
	case *types.Alias:
		return g.wrap(types.Unalias(x))
	case *types.Array:
		return fmt.Sprintf("[%d]%s", x.Len(), g.wrap(x.Elem()))
	case *types.Chan:
		return fmt.Sprintf("%s%s", NewChanDir(x.Dir()), g.wrap(x.Elem()))
	case *types.Map:
		return fmt.Sprintf("map[%s]%s", g.wrap(x.Key()), g.wrap(x.Elem()))
	case *types.Interface:
		b := strings.Builder{}
		b.WriteString("interface { ")
		for i := range x.NumMethods() {
			if i > 0 {
				b.WriteString("; ")
			}
			m := x.Method(i)
			f := g.wrap(m.Signature())
			b.WriteString(m.Name() + f[4:])
		}
		b.WriteString(" }")
		return b.String()
	case *types.Pointer:
		return fmt.Sprintf("*%s", g.wrap(x.Elem()))
	case *types.Signature:
		b := strings.Builder{}
		b.WriteString("func(")
		for i := range x.Params().Len() {
			if i > 0 {
				b.WriteString(", ")
			}
			p := x.Params().At(i)
			if x.Variadic() && i == x.Params().Len()-1 {
				b.WriteString("...")
				b.WriteString(g.wrap(p.Type().(*types.Slice).Elem()))
				break
			}
			b.WriteString(g.wrap(p.Type()))
		}
		// params := make([]*types.Var, 0, x.Params().Len()+1)
		// if x.Recv() != nil {
		// 	params = append(params, x.Recv())
		// }
		// for i := range x.Params().Len() {
		// 	params = append(params, x.Params().At(i))
		// }
		// for i := range params {
		// 	if i > 0 {
		// 		b.WriteString(", ")
		// 	}
		// 	p := params[i]
		// 	if x.Variadic() && i == x.Params().Len()-1 {
		// 		b.WriteString("...")
		// 		b.WriteString(g.wrap(p.Type().(*types.Slice).Elem()))
		// 		break
		// 	}
		// 	b.WriteString(g.wrap(p.Type()))
		// }
		b.WriteString(")")
		if x.Results().Len() == 0 {
			return b.String()
		}
		b.WriteString(" ")
		if x.Results().Len() == 1 {
			b.WriteString(g.wrap(x.Results().At(0).Type()))
			return b.String()
		}
		b.WriteString("(")
		for i := range x.Results().Len() {
			if i > 0 {
				b.WriteString(", ")
			}
			r := x.Results().At(i)
			b.WriteString(g.wrap(r.Type()))
		}
		b.WriteString(")")

		return b.String()
	case *types.Slice:
		return fmt.Sprintf("[]%s", g.wrap(x.Elem()))
	case *types.Struct:
		if x.NumFields() == 0 {
			return "struct {}"
		}
		b := strings.Builder{}
		b.WriteString("struct { ")
		for i := range x.NumFields() {
			if i > 0 {
				b.WriteString("; ")
			}
			f := x.Field(i)
			if !f.Anonymous() {
				b.WriteString(f.Name())
				b.WriteString(" ")
			}
			b.WriteString(g.wrap(f.Type()))
			if tag := x.Tag(i); len(tag) > 0 {
				b.WriteString(" ")
				b.WriteString(strconv.Quote(tag))
			}
		}
		b.WriteString(" }")
		return b.String()
	default:
		n, ok := t.(*types.Named)
		must.BeTrueF(ok, "invalid WrapT type: %T", x)
		ok = n.TypeArgs().Len() == n.TypeParams().Len()
		must.BeTrueF(ok, "uninstantiated generic type: %s", x.String())
		ok = n.Obj().Pkg() != nil
		must.BeTrueF(ok, "unexpect nil package type: %s", x.String())
		b := strings.Builder{}
		b.WriteString(n.Obj().Pkg().Path())
		b.WriteString(".")
		b.WriteString(n.Obj().Name())
		if n.TypeArgs().Len() > 0 {
			b.WriteString("[")
			for i := range n.TypeArgs().Len() {
				if i > 0 {
					b.WriteString(",")
				}
				targ := n.TypeArgs().At(i)
				b.WriteString(g.wrap(targ))
			}
			b.WriteString("]")
		}
		return wrap(b.String())
	}
}
