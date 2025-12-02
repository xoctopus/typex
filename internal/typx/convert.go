package typx

import (
	"go/types"
	"reflect"

	"github.com/xoctopus/x/misc/must"
)

func NewTTByLit(t *LitType) types.Type {
	if t.typename != "" {
		if x, ok := gTBasicKinds.Load(t.typename); ok && t.pkg == "" {
			return x
		}
		if t.pkg == "unsafe" && t.typename == "Pointer" {
			return types.Typ[types.UnsafePointer]
		}

		must.BeTrue(t.PkgPath() != "")
		typ := Lookup[*types.Named](Load(t.PkgPath()), t.typename)
		must.BeTrue(typ != nil)
		must.BeTrue(typ.TypeParams().Len() == len(t.targs))
		if len(t.targs) > 0 {
			args := make([]types.Type, len(t.targs))
			for i, arg := range t.targs {
				args[i] = NewTTByLit(arg)
			}
			typ = Instantiate(typ, args...).(*types.Named)
		}
		return typ
	}

	switch t.kind {
	case reflect.Array:
		return types.NewArray(NewTTByLit(t.ele), int64(t.len))
	case reflect.Chan:
		return types.NewChan(TChanDir(t.dir), NewTTByLit(t.ele))
	case reflect.Func:
		ins := make([]*types.Var, len(t.ins))
		for i, v := range t.ins {
			pkg := (*types.Package)(nil)
			if path := v.PkgPath(); path != "" {
				pkg = types.NewPackage(path, "")
			}
			ins[i] = types.NewParam(0, pkg, "", NewTTByLit(v))
		}
		outs := make([]*types.Var, len(t.outs))
		for i, v := range t.outs {
			pkg := (*types.Package)(nil)
			if path := v.PkgPath(); path != "" {
				pkg = types.NewPackage(path, "")
			}
			outs[i] = types.NewParam(0, pkg, "", NewTTByLit(v))
		}
		return types.NewSignatureType(
			nil, nil, nil,
			types.NewTuple(ins...), types.NewTuple(outs...),
			t.variadic,
		)
	case reflect.Interface:
		methods := make([]*types.Func, len(t.methods))
		for i, m := range t.methods {
			s := NewTTByLit(m).(*types.Signature)
			methods[i] = types.NewFunc(0, nil, m.name, s)
		}
		return types.NewInterfaceType(methods, nil)
	case reflect.Map:
		return types.NewMap(NewTTByLit(t.key), NewTTByLit(t.ele))
	case reflect.Pointer:
		return types.NewPointer(NewTTByLit(t.ele))
	case reflect.Slice:
		return types.NewSlice(NewTTByLit(t.ele))
	default:
		must.BeTrueF(t.kind == reflect.Struct, "unexpected kind %s", t.kind)
		fields := make([]*types.Var, len(t.fields))
		tags := make([]string, len(t.fields))
		for i, f := range t.fields {
			pkg := (*types.Package)(nil)
			if path := f.PkgPath(); path != "" {
				pkg = types.NewPackage(path, "")
			}
			fields[i] = types.NewField(0, pkg, f.name, NewTTByLit(f), f.embedded)
			tags[i] = f.tag
		}
		return types.NewStruct(fields, tags)
	}
}

// NewTTByRT parse reflect.Type to typx.LitType and instantiate a types.Type by
// package scan. make sure the package context is unique or std only.
func NewTTByRT(r reflect.Type) types.Type {
	return NewTTByLit(NewLitTypeByID(wrapRT(r)))
}
