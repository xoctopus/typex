package internal

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/resultx"
	"github.com/xoctopus/x/stringsx"

	"github.com/xoctopus/typex/internal/gtypex"
	"github.com/xoctopus/typex/namer"
	"github.com/xoctopus/typex/pkgutil"
)

type Literal interface {
	// PkgPath returns type's full package path
	PkgPath() string
	// Name returns type's name with type arguments
	Name() string
	// String return type's string with full package path everywhere
	String() string
	// TType returns type's types.Type
	TType() types.Type
	// TypeLit returns type's literal, it should be consistent with the literal
	// representation shown in source code
	TypeLit(context.Context) string
}

func literalize(id string) Literal {
	id = g.wrap(id)
	ident := func(code string, x ast.Node) string {
		return code[x.Pos()-1 : x.End()-1]
	}

	expr := resultx.Unwrap(parser.ParseExpr(id))
	switch e := expr.(type) {
	case *ast.ArrayType:
		if e.Len != nil {
			return utype{
				kind: reflect.Array,
				elem: g.literalize(ident(id, e.Elt)),
				len:  resultx.Unwrap(stringsx.Atoi(ident(id, e.Len))),
			}
		}
		return utype{
			kind: reflect.Slice,
			elem: g.literalize(ident(id, e.Elt)),
		}
	case *ast.ChanType:
		return utype{
			kind: reflect.Chan,
			dir:  NewChanDir(e.Dir),
			elem: g.literalize(ident(id, e.Value)),
		}
	case *ast.FuncType:
		u := utype{kind: reflect.Func}
		if e.Params != nil && len(e.Params.List) > 0 {
			u.ins = make([]Literal, len(e.Params.List))
			for i, p := range e.Params.List {
				param := ident(id, p.Type)
				if i == len(e.Params.List)-1 && strings.HasPrefix(param, "...") {
					u.variadic = true
					u.ins[i] = g.literalize("[]" + param[3:])
					break
				}
				u.ins[i] = g.literalize(param)
			}
		}
		if e.Results != nil && len(e.Results.List) > 0 {
			u.outs = make([]Literal, len(e.Results.List))
			for i, r := range e.Results.List {
				u.outs[i] = g.literalize(ident(id, r.Type))
			}
		}
		return u
	case *ast.InterfaceType:
		u := utype{
			kind:    reflect.Interface,
			methods: make([]Literal, len(e.Methods.List)),
		}
		for i, m := range e.Methods.List {
			mi := g.literalize("func" + ident(id, m.Type)).(utype)
			mi.name = m.Names[0].Name
			u.methods[i] = mi
		}
		return u
	case *ast.MapType:
		return utype{
			kind: reflect.Map,
			key:  g.literalize(ident(id, e.Key)),
			elem: g.literalize(ident(id, e.Value)),
		}
	case *ast.StarExpr:
		return utype{
			kind: reflect.Pointer,
			elem: g.literalize(ident(id, e.X)),
		}
	case *ast.StructType:
		u := utype{kind: reflect.Struct}
		if e.Fields != nil {
			u.fields = make([]*ufield, len(e.Fields.List))
			for i, f := range e.Fields.List {
				uf := &ufield{typ: g.literalize(ident(id, f.Type))}
				if uf.embedded = len(f.Names) == 0; uf.embedded {
					uf.name = uf.typ.Name()
				} else {
					uf.name = f.Names[0].Name
				}
				if f.Tag != nil {
					uf.tag = resultx.Unwrap(strconv.Unquote(f.Tag.Value))
				}
				u.fields[i] = uf
			}
		}
		return u
	// ident will treat as builtin type
	// case *ast.Ident:
	// 	return utype{typename: e.Name}
	case *ast.SelectorExpr:
		return utype{
			pkg:      pkgutil.New(ident(id, e.X)),
			typename: ident(id, e.Sel),
		}
	case *ast.IndexExpr:
		u := g.literalize(ident(id, e.X)).(utype)
		u.targs = []Literal{g.literalize(ident(id, e.Index))}
		return u
	default:
		ex, ok := e.(*ast.IndexListExpr)
		must.BeTrueF(ok, "unexpected expr [%T] %s", e, ident(id, e))
		u := g.literalize(ident(id, ex.X)).(utype)
		u.targs = make([]Literal, len(ex.Indices))
		for i, index := range ex.Indices {
			u.targs[i] = g.literalize(ident(id, index))
		}
		return u
	}
}

func literalizeRT(t reflect.Type) Literal {
	if id := t.Name(); id != "" {
		must.BeTrue(t.PkgPath() != "")
		return g.literalize(t.PkgPath() + "." + id)
	}

	u := utype{kind: t.Kind()}
	switch t.Kind() {
	case reflect.Array:
		u.len, u.elem = t.Len(), g.literalize(t.Elem())
	case reflect.Chan:
		u.dir, u.elem = NewChanDir(t.ChanDir()), g.literalize(t.Elem())
	case reflect.Func:
		u.variadic = t.IsVariadic()
		if n := t.NumIn(); n > 0 {
			u.ins = make([]Literal, n)
			for i := range u.ins {
				u.ins[i] = g.literalize(t.In(i))
			}
		}
		if n := t.NumOut(); n > 0 {
			u.outs = make([]Literal, n)
			for i := range u.outs {
				u.outs[i] = g.literalize(t.Out(i))
			}
		}
	case reflect.Interface:
		u.methods = make([]Literal, t.NumMethod())
		for i := range u.methods {
			m := t.Method(i)
			mi := g.literalize(m.Type).(utype)
			mi.name = m.Name
			u.methods[i] = mi
		}
	case reflect.Map:
		u.key, u.elem = g.literalize(t.Key()), g.literalize(t.Elem())
	case reflect.Pointer:
		u.elem = g.literalize(t.Elem())
	case reflect.Slice:
		u.elem = g.literalize(t.Elem())
	default:
		must.BeTrue(t.Kind() == reflect.Struct)
		u.fields = make([]*ufield, t.NumField())
		for i := range t.NumField() {
			f := t.Field(i)
			u.fields[i] = &ufield{
				name:     f.Name,
				typ:      g.literalize(f.Type),
				tag:      string(f.Tag),
				embedded: f.Anonymous,
			}
		}
	}
	return u
}

func literalizeTT(t types.Type) Literal {
	switch x := t.(type) {
	case *types.Alias:
		return g.literalize(types.Unalias(x))
	case *types.Array:
		return utype{
			kind: reflect.Array,
			len:  int(x.Len()),
			elem: g.literalize(x.Elem()),
		}
	case *types.Chan:
		return utype{
			kind: reflect.Chan,
			dir:  NewChanDir(x.Dir()),
			elem: g.literalize(x.Elem()),
		}
	case *types.Map:
		return utype{
			kind: reflect.Map,
			key:  g.literalize(x.Key()),
			elem: g.literalize(x.Elem()),
		}
	case *types.Interface:
		methods := make([]Literal, x.NumMethods())
		for i := range methods {
			m := x.Method(i)
			mi := g.literalize(m.Signature()).(utype)
			mi.name = m.Name()
			methods[i] = mi
		}
		return utype{
			kind:    reflect.Interface,
			methods: methods,
		}
	case *types.Pointer:
		return utype{
			kind: reflect.Pointer,
			elem: g.literalize(x.Elem()),
		}
	case *types.Signature:
		u := utype{
			kind:     reflect.Func,
			variadic: x.Variadic(),
		}
		if n := x.Params().Len(); n > 0 {
			u.ins = make([]Literal, n)
			for i := range u.ins {
				u.ins[i] = g.literalize(x.Params().At(i).Type())
			}
		}
		if n := x.Results().Len(); n > 0 {
			u.outs = make([]Literal, n)
			for i := range u.outs {
				u.outs[i] = g.literalize(x.Results().At(i).Type())
			}
		}
		return u
	case *types.Slice:
		return utype{
			kind: reflect.Slice,
			elem: g.literalize(x.Elem()),
		}
	case *types.Struct:
		fields := make([]*ufield, x.NumFields())
		for i := range fields {
			f := x.Field(i)
			ff := &ufield{
				typ:      g.literalize(f.Type()),
				name:     f.Name(),
				tag:      x.Tag(i),
				embedded: f.Embedded(),
			}
			if ff.embedded {
				ff.name = ff.typ.Name()
			}
			fields[i] = ff
		}
		return utype{
			kind:   reflect.Struct,
			fields: fields,
		}
	default:
		xx, ok := t.(*types.Named)
		must.BeTrueF(ok, "")
		u := utype{
			pkg:      pkgutil.NewT(xx.Obj().Pkg()),
			typename: xx.Obj().Name(),
		}
		if xx.TypeArgs().Len() > 0 {
			u.targs = make([]Literal, xx.TypeArgs().Len())
			for i := range u.targs {
				u.targs[i] = g.literalize(xx.TypeArgs().At(i))
			}
		}
		return u
	}
}

type utype struct {
	pkg      pkgutil.Package
	typename string
	targs    []Literal
	kind     reflect.Kind
	name     string
	key      Literal
	elem     Literal
	len      int
	dir      ChanDir
	ins      []Literal
	outs     []Literal
	variadic bool
	fields   []*ufield
	methods  []Literal
}

func (t utype) PkgPath() string {
	if t.pkg != nil {
		return t.pkg.Path()
	}
	return ""
}

func (t utype) Name() string {
	if t.typename != "" {
		b := strings.Builder{}
		b.WriteString(t.typename)
		if len(t.targs) > 0 {
			b.WriteString("[")
			for i, targ := range t.targs {
				if i > 0 {
					b.WriteString(",")
				}
				b.WriteString(targ.String())
			}
			b.WriteString("]")
		}
		return b.String()
	}
	return ""
}

func (t utype) TypeLit(ctx context.Context) string {
	if t.pkg != nil {
		must.BeTrue(t.typename != "")

		name := t.pkg.Name()
		pkgnamer, _ := namer.FromContext(ctx)
		if pkgnamer != nil {
			name = pkgnamer.Package(t.pkg.Path())
		}

		b := strings.Builder{}
		if name != "" {
			b.WriteString(name)
			b.WriteString(".")
		}
		b.WriteString(t.typename)

		if len(t.targs) == 0 {
			return b.String()
		}
		b.WriteString("[")
		for i, targ := range t.targs {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(targ.TypeLit(ctx))
		}
		b.WriteString("]")
		return b.String()
	}

	switch t.kind {
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.len, t.elem.TypeLit(ctx))
	case reflect.Chan:
		return fmt.Sprintf("%s%s", t.dir.String(), t.elem.TypeLit(ctx))
	case reflect.Func:
		b := strings.Builder{}

		name := "func"
		if t.name != "" {
			name = t.name
		}
		b.WriteString(name + "(")
		for i := range t.ins {
			if i > 0 {
				b.WriteString(", ")
			}
			if i == len(t.ins)-1 && t.variadic {
				b.WriteString("..." + t.ins[i].TypeLit(ctx)[2:])
				break
			}
			b.WriteString(t.ins[i].TypeLit(ctx))
		}
		b.WriteString(")")

		if len(t.outs) == 0 {
			return b.String()
		}
		b.WriteString(" ")
		if len(t.outs) > 1 {
			b.WriteString("(")
		}
		for i, v := range t.outs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(v.TypeLit(ctx))
		}
		if len(t.outs) > 1 {
			b.WriteString(")")
		}
		return b.String()
	case reflect.Interface:
		b := strings.Builder{}
		b.WriteString("interface { ")
		for i, m := range t.methods {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(m.TypeLit(ctx))
		}
		b.WriteString(" }")
		return b.String()
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", t.key.TypeLit(ctx), t.elem.TypeLit(ctx))
	case reflect.Pointer:
		return fmt.Sprintf("*%s", t.elem.TypeLit(ctx))
	case reflect.Slice:
		return fmt.Sprintf("[]%s", t.elem.TypeLit(ctx))
	default:
		must.BeTrue(t.kind == reflect.Struct)
		if len(t.fields) == 0 {
			return "struct {}"
		}
		b := strings.Builder{}
		b.WriteString("struct { ")
		for i, f := range t.fields {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(f.TypeLit(ctx))
		}
		b.WriteString(" }")
		return b.String()
	}
}

func (t utype) String() string {
	if t.typename != "" {
		b := strings.Builder{}
		if t.pkg != nil {
			b.WriteString(t.pkg.Path())
			b.WriteString(".")
		}
		b.WriteString(t.Name())
		return b.String()
	}

	switch t.kind {
	case reflect.Array:
		return "[" + strconv.Itoa(t.len) + "]" + t.elem.String()
	case reflect.Chan:
		return t.dir.String() + t.elem.String()
	case reflect.Func:
		b := strings.Builder{}

		name := "func"
		if t.name != "" {
			name = t.name
		}
		b.WriteString(name + "(")
		for i := range t.ins {
			if i > 0 {
				b.WriteString(", ")
			}
			if i == len(t.ins)-1 && t.variadic {
				b.WriteString("..." + t.ins[i].String()[2:])
				break
			}
			b.WriteString(t.ins[i].String())
		}
		b.WriteString(")")

		if len(t.outs) == 0 {
			return b.String()
		}
		b.WriteString(" ")
		if len(t.outs) > 1 {
			b.WriteString("(")
		}
		for i, v := range t.outs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(v.String())
		}
		if len(t.outs) > 1 {
			b.WriteString(")")
		}
		return b.String()
	case reflect.Interface:
		// empty interface will treat as a builtin type
		// if len(t.methods) == 0 {
		// 	return "interface {}"
		// }
		b := strings.Builder{}
		b.WriteString("interface { ")
		for i, m := range t.methods {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(m.String())
		}
		b.WriteString(" }")
		return b.String()
	case reflect.Map:
		return "map[" + t.key.String() + "]" + t.elem.String()
	case reflect.Pointer:
		return "*" + t.elem.String()
	case reflect.Slice:
		return "[]" + t.elem.String()
	default:
		must.BeTrue(t.kind == reflect.Struct)
		if len(t.fields) == 0 {
			return "struct {}"
		}
		b := strings.Builder{}
		b.WriteString("struct { ")
		for i, f := range t.fields {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(f.String())
		}
		b.WriteString(" }")
		return b.String()
	}
}

func (t utype) TType() types.Type {
	switch t.kind {
	case reflect.Array:
		return types.NewArray(t.elem.TType(), int64(t.len))
	case reflect.Chan:
		return types.NewChan(t.dir.TypesChanDir(), t.elem.TType())
	case reflect.Func:
		ins := make([]*types.Var, len(t.ins))
		for i, v := range t.ins {
			var pkg *types.Package
			if p := pkgutil.New(v.PkgPath()); p != nil {
				pkg = p.Unwrap()
			}
			ins[i] = types.NewParam(0, pkg, "", v.TType())
		}
		outs := make([]*types.Var, len(t.outs))
		for i, v := range t.outs {
			var pkg *types.Package
			if p := pkgutil.New(v.PkgPath()); p != nil {
				pkg = p.Unwrap()
			}
			outs[i] = types.NewParam(0, pkg, "", v.TType())
		}
		return types.NewSignatureType(
			nil, nil, nil,
			types.NewTuple(ins...), types.NewTuple(outs...),
			t.variadic,
		)
	case reflect.Interface:
		methods := make([]*types.Func, len(t.methods))
		for i, m := range t.methods {
			mm := m.(utype)
			s := mm.TType().(*types.Signature)
			methods[i] = types.NewFunc(0, nil, mm.name, s)
		}
		return types.NewInterfaceType(methods, nil)
	case reflect.Map:
		return types.NewMap(t.key.TType(), t.elem.TType())
	case reflect.Pointer:
		return types.NewPointer(t.elem.TType())
	case reflect.Slice:
		return types.NewSlice(t.elem.TType())
	case reflect.Struct:
		fields := make([]*types.Var, len(t.fields))
		tags := make([]string, len(t.fields))
		for i, f := range t.fields {
			pkg := (*types.Package)(nil)
			if p := pkgutil.New(f.typ.PkgPath()); p != nil {
				pkg = p.Unwrap()
			}
			fields[i] = types.NewField(0, pkg, f.name, f.typ.TType(), f.embedded)
			tags[i] = f.tag
		}
		return types.NewStruct(fields, tags)
	default:
		must.BeTrue(t.pkg != nil && t.typename != "")
		typ := pkgutil.MustLookup[*types.Named](t.pkg, t.typename)
		must.BeTrue(typ != nil)
		must.BeTrue(typ.TypeParams().Len() == len(t.targs))
		if len(t.targs) > 0 {
			args := make([]types.Type, len(t.targs))
			for i, arg := range t.targs {
				args[i] = arg.TType()
			}
			typ = gtypex.Instantiate(typ, args).(*types.Named)
		}
		return typ
	}
}

type ufield struct {
	name     string
	typ      Literal
	tag      string
	embedded bool
}

func (v ufield) TypeLit(ctx context.Context) string {
	b := strings.Builder{}
	if !v.embedded {
		b.WriteString(v.name)
		b.WriteString(" ")
	}
	b.WriteString(v.typ.TypeLit(ctx))
	if len(v.tag) > 0 {
		b.WriteString(" ")
		b.WriteString(strconv.Quote(v.tag))
	}
	return b.String()
}

func (v ufield) String() string {
	b := strings.Builder{}
	if !v.embedded {
		b.WriteString(v.name)
		b.WriteString(" ")
	}
	b.WriteString(v.typ.String())
	if len(v.tag) > 0 {
		b.WriteString(" ")
		b.WriteString(strconv.Quote(v.tag))
	}
	return b.String()
}
