package typx

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/stringsx"
)

// type LitType interface {
// 	// Underlying returns literal type's underlying, it must be `reflect.Type` or `types.Type`
// 	Underlying() any
// 	// PkgPath returns type's full package path
// 	PkgPath() string
// 	// Name returns type's name with type arguments
// 	Name() string
// 	// String return type's string with full package path everywhere
// 	String() string
// 	// Type returns type's types.Type
// 	// Type() types.Type
// 	// Lit returns type's literal, it should be consistent with the literal
// 	// representation shown in source code
// 	// Lit(context.Context) string
//
// 	// literal returns type string. if w == true returns wrapped string
// 	literal(bool) string
// }

func NewLitType(t any) (x *LitType) {
	switch u := t.(type) {
	case reflect.Type:
		if l, ok := gRLiterals.Load(u); ok {
			return l
		}
	case types.Type:
		if l, ok := gTLiterals.Load(u); ok {
			return l
		}
	default:
		panic(fmt.Errorf("unexpect input for new a LitType from %T", u))
	}

	defer func() {
		if x != nil {
			x.underlying = t
			switch u := t.(type) {
			case reflect.Type:
				gRLiterals.Store(u, x)
			case types.Type:
				gTLiterals.Store(u, x)
			}
		}
	}()

	id := Wrap(t)
	return NewTypeByID(id)
}

func NewTypeByID(id string) (x *LitType) {
	ident := func(code string, x ast.Node) string {
		return code[x.Pos()-1 : x.End()-1]
	}
	expr := must.NoErrorV(parser.ParseExpr(id))

	switch e := expr.(type) {
	case *ast.ArrayType:
		if e.Len != nil {
			return &LitType{
				kind: reflect.Array,
				ele:  NewTypeByID(ident(id, e.Elt)),
				len:  must.NoErrorV(stringsx.Atoi(ident(id, e.Len))),
			}
		}
		return &LitType{
			kind: reflect.Slice,
			ele:  NewTypeByID(ident(id, e.Elt)),
		}
	case *ast.ChanType:
		return &LitType{
			kind: reflect.Chan,
			dir:  e.Dir,
			ele:  NewTypeByID(ident(id, e.Value)),
		}
	case *ast.FuncType:
		u := &LitType{kind: reflect.Func}
		if e.Params != nil && len(e.Params.List) > 0 {
			u.ins = make([]*LitType, len(e.Params.List))
			for i, p := range e.Params.List {
				param := ident(id, p.Type)
				if i == len(e.Params.List)-1 && strings.HasPrefix(param, "...") {
					u.variadic = true
					u.ins[i] = NewTypeByID("[]" + param[3:])
					break
				}
				u.ins[i] = NewTypeByID(param)
			}
		}
		if e.Results != nil && len(e.Results.List) > 0 {
			u.outs = make([]*LitType, len(e.Results.List))
			for i, r := range e.Results.List {
				u.outs[i] = NewTypeByID(ident(id, r.Type))
			}
		}
		return u
	case *ast.InterfaceType:
		u := &LitType{
			kind:    reflect.Interface,
			methods: make([]*LitType, len(e.Methods.List)),
		}
		for i, m := range e.Methods.List {
			mi := NewTypeByID("func" + ident(id, m.Type))
			mi.name = m.Names[0].Name
			u.methods[i] = mi
		}
		return u
	case *ast.MapType:
		return &LitType{
			kind: reflect.Map,
			key:  NewTypeByID(ident(id, e.Key)),
			ele:  NewTypeByID(ident(id, e.Value)),
		}
	case *ast.StarExpr:
		return &LitType{
			kind: reflect.Pointer,
			ele:  NewTypeByID(ident(id, e.X)),
		}
	case *ast.StructType:
		u := &LitType{kind: reflect.Struct}
		if e.Fields != nil {
			u.fields = make([]*LitType, 0, len(e.Fields.List))

			for _, f := range e.Fields.List {
				ft := NewTypeByID(ident(id, f.Type))
				if f.Tag != nil {
					ft.tag = must.NoErrorV(strconv.Unquote(f.Tag.Value))
				}
				if len(f.Names) == 0 {
					ft.name = ft.Name()
					if idx := strings.Index(ft.name, "["); idx != -1 {
						ft.name = ft.name[:idx]
					}
					ft.embedded = true
					u.fields = append(u.fields, ft)
				} else {
					for _, n := range f.Names {
						_ft := *ft
						_ft.name = n.Name
						u.fields = append(u.fields, &_ft)
					}
				}
			}
		}
		return u
	case *ast.SelectorExpr:
		u := &LitType{
			pkg:      ident(id, e.X),
			typename: ident(id, e.Sel),
		}
		if u.pkg == "unsafe" && u.typename == "Pointer" {
			u.kind = reflect.UnsafePointer
		}
		return u
	case *ast.IndexExpr:
		u := NewTypeByID(ident(id, e.X))
		u.targs = []*LitType{NewTypeByID(ident(id, e.Index))}
		return u
	case *ast.IndexListExpr:
		u := NewTypeByID(ident(id, e.X))
		u.targs = make([]*LitType, len(e.Indices))
		for i, index := range e.Indices {
			u.targs[i] = NewTypeByID(ident(id, index))
		}
		return u
	default:
		code := ident(id, e)
		ex, ok := e.(*ast.Ident)
		must.BeTrueF(ok, "expect an ast.Ident but caught %T: %s", ex, code)
		u := &LitType{typename: ex.Name}
		if k, ok := gRBasicKinds.Load(ex.Name); ok {
			u.kind = k
		}
		return u
	}
}

type LitType struct {
	underlying any
	pkg        string
	typename   string
	targs      []*LitType
	kind       reflect.Kind
	name       string
	key        *LitType
	ele        *LitType
	len        int
	dir        any // ChanDir
	ins        []*LitType
	outs       []*LitType
	variadic   bool
	fields     []*LitType
	methods    []*LitType
	tag        string
	embedded   bool
}

func (t *LitType) Underlying() any {
	return t.underlying
}

func (t *LitType) PkgPath() string {
	return DecodePath(t.pkg)
}

func (t *LitType) Name() string {
	if t.typename == "" || len(t.targs) == 0 {
		return t.typename
	}

	b := strings.Builder{}
	b.WriteString(t.typename)
	b.WriteRune('[')
	for i, targ := range t.targs {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(targ.String())
	}
	b.WriteRune(']')
	return b.String()
}

func (t *LitType) Kind() reflect.Kind {
	return t.kind
}

func (t *LitType) literal(w bool) string {
	if t.typename != "" {
		b := strings.Builder{}
		if path := t.pkg; path != "" {
			if !w {
				path = DecodePath(path) // origin
			} else {
				path = EncodePath(path) // wrapped
			}
			b.WriteString(path)
			b.WriteString(".")
		}
		b.WriteString(t.typename)
		if len(t.targs) > 0 {
			b.WriteRune('[')
			for i, targ := range t.targs {
				if i > 0 {
					b.WriteString(",")
				}
				b.WriteString(targ.literal(w))
			}
			b.WriteRune(']')
		}
		return b.String()
	}

	switch t.kind {
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.len, t.ele.literal(w))
	case reflect.Chan:
		return fmt.Sprintf("%s%s", ChanDir(t.dir), t.ele.literal(w))
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
				b.WriteString("..." + t.ins[i].literal(w)[2:])
				break
			}
			b.WriteString(t.ins[i].literal(w))
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
			b.WriteString(v.literal(w))
		}
		if len(t.outs) > 1 {
			b.WriteString(")")
		}
		return b.String()
	case reflect.Interface:
		if len(t.methods) == 0 {
			return "interface {}"
		}
		b := strings.Builder{}
		b.WriteString("interface { ")
		for i, m := range t.methods {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(m.literal(w))
		}
		b.WriteString(" }")
		return b.String()
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", t.key.literal(w), t.ele.literal(w))
	case reflect.Pointer:
		return "*" + t.ele.literal(w)
	case reflect.Slice:
		return "[]" + t.ele.literal(w)
	default:
		must.BeTrueF(t.kind == reflect.Struct, "got unexpected kind %s", t.kind)
		if len(t.fields) == 0 {
			return "struct {}"
		}
		b := strings.Builder{}
		b.WriteString("struct { ")
		for i, f := range t.fields {
			if i > 0 {
				b.WriteString("; ")
			}
			if !f.embedded {
				b.WriteString(f.name)
				b.WriteString(" ")
			}
			b.WriteString(f.literal(w))
			if len(f.tag) > 0 {
				b.WriteString(" ")
				b.WriteString(strconv.Quote(f.tag))
			}
		}
		b.WriteString(" }")
		return b.String()
	}
}

func (t *LitType) Literal() string {
	return t.literal(true)
}

func (t *LitType) String() string {
	return t.literal(false)
}

func (t *LitType) Type() (x types.Type) {
	if tt, ok := t.underlying.(types.Type); ok {
		return tt
	}

	if t.typename != "" {
		if x, ok := gTBasicKinds.Load(t.typename); ok && t.pkg == "" {
			return x
		}
		if t.pkg == "unsafe" && t.typename == "Pointer" {
			return types.Typ[types.UnsafePointer]
		}

		must.BeTrue(t.pkg != "")
		typ := Lookup[*types.Named](Load(t.PkgPath()), t.typename)
		must.BeTrue(typ != nil)
		must.BeTrue(typ.TypeParams().Len() == len(t.targs))
		if len(t.targs) > 0 {
			args := make([]types.Type, len(t.targs))
			for i, arg := range t.targs {
				args[i] = arg.Type()
			}
			typ = Instantiate(typ, args...).(*types.Named)
		}
		return typ
	}

	switch t.kind {
	case reflect.Array:
		return types.NewArray(t.ele.Type(), int64(t.len))
	case reflect.Chan:
		return types.NewChan(TChanDir(t.dir), t.ele.Type())
	case reflect.Func:
		ins := make([]*types.Var, len(t.ins))
		for i, v := range t.ins {
			pkg := (*types.Package)(nil)
			if path := v.PkgPath(); path != "" {
				pkg = types.NewPackage(path, "")
			}
			ins[i] = types.NewParam(0, pkg, "", v.Type())
		}
		outs := make([]*types.Var, len(t.outs))
		for i, v := range t.outs {
			pkg := (*types.Package)(nil)
			if path := v.PkgPath(); path != "" {
				pkg = types.NewPackage(path, "")
			}
			outs[i] = types.NewParam(0, pkg, "", v.Type())
		}
		return types.NewSignatureType(
			nil, nil, nil,
			types.NewTuple(ins...), types.NewTuple(outs...),
			t.variadic,
		)
	case reflect.Interface:
		methods := make([]*types.Func, len(t.methods))
		for i, m := range t.methods {
			s := m.Type().(*types.Signature)
			methods[i] = types.NewFunc(0, nil, m.name, s)
		}
		return types.NewInterfaceType(methods, nil)
	case reflect.Map:
		return types.NewMap(t.key.Type(), t.ele.Type())
	case reflect.Pointer:
		return types.NewPointer(t.ele.Type())
	case reflect.Slice:
		return types.NewSlice(t.ele.Type())
	default:
		must.BeTrueF(t.kind == reflect.Struct, "unexpected kind %s", t.kind)
		fields := make([]*types.Var, len(t.fields))
		tags := make([]string, len(t.fields))
		for i, f := range t.fields {
			pkg := (*types.Package)(nil)
			if path := f.PkgPath(); path != "" {
				pkg = types.NewPackage(path, "")
			}
			fields[i] = types.NewField(0, pkg, f.name, f.Type(), f.embedded)
			tags[i] = f.tag
		}
		return types.NewStruct(fields, tags)
	}
}
