package gtypex

import (
	"go/types"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/resultx"
)

type GenericType interface {
	TypeParams() *types.TypeParamList
	TypeArgs() *types.TypeList
}

type InterfaceOrNamed interface {
	NumMethods() int
	Method(int) *types.Func
}

func Instantiate(t types.Type, args []types.Type) types.Type {
	must.BeTrue(len(args) > 0)

	switch x := t.(type) {
	case *types.Alias:
		return Instantiate(types.Unalias(x), args)
	case *types.Array:
		return types.NewArray(Instantiate(x.Elem(), args), x.Len())
	case *types.Basic:
		return t
	case *types.Chan:
		return types.NewChan(x.Dir(), Instantiate(x.Elem(), args))
	case *types.Interface:
		methods := make([]*types.Func, x.NumMethods())
		for i := range x.NumMethods() {
			m := x.Method(i)
			s := Instantiate(m.Signature(), args).(*types.Signature)
			methods[i] = types.NewFunc(0, m.Pkg(), m.Name(), s)
		}
		return types.NewInterfaceType(methods, nil)
	case *types.Map:
		return types.NewMap(Instantiate(x.Key(), args), Instantiate(x.Elem(), args))
	case *types.Pointer:
		return types.NewPointer(Instantiate(x.Elem(), args))
	case *types.Slice:
		return types.NewSlice(Instantiate(x.Elem(), args))
	case *types.Signature:
		return types.NewSignatureType(
			nil, nil, nil,
			Instantiate(x.Params(), args).(*types.Tuple),
			Instantiate(x.Results(), args).(*types.Tuple),
			x.Variadic(),
		)
	case *types.Struct:
		fields := make([]*types.Var, x.NumFields())
		tags := make([]string, x.NumFields())
		for i := range x.NumFields() {
			v := x.Field(i)
			fields[i] = types.NewField(0, v.Pkg(), v.Name(), Instantiate(v.Type(), args), v.Anonymous())
			tags[i] = x.Tag(i)
		}
		return types.NewStruct(fields, tags)
	case *types.TypeParam:
		return args[x.Index()]
	case *types.Tuple:
		vars := make([]*types.Var, x.Len())
		for i := range x.Len() {
			v := x.At(i)
			vars[i] = types.NewParam(0, v.Pkg(), v.Name(), Instantiate(v.Type(), args))
		}
		return types.NewTuple(vars...)
	default:
		n := t.(*types.Named)
		if n.TypeParams().Len() == 0 {
			return n
		}
		targs := make([]types.Type, n.TypeParams().Len())
		if argc := n.TypeArgs().Len(); argc > 0 {
			must.BeTrue(argc == n.TypeParams().Len())
			for i := range argc {
				if p, ok := n.TypeArgs().At(i).(*types.TypeParam); ok {
					targs[i] = Instantiate(p, args)
				} else {
					targs[i] = n.TypeArgs().At(i)
				}
			}
		} else {
			must.BeTrue(n.TypeParams().Len() == len(args))
			targs = args
		}
		return resultx.Unwrap(types.Instantiate(nil, x, targs, true))
	}
}

// Underlying returns t's instantiated underlying type.
func Underlying(t types.Type) types.Type {
	tt, ok := t.(GenericType)
	if !ok || tt.TypeParams().Len() == 0 {
		return t.Underlying()
	}

	must.BeTrue(tt.TypeParams().Len() == tt.TypeArgs().Len())
	args := make([]types.Type, tt.TypeArgs().Len())
	for i := range args {
		args[i] = tt.TypeArgs().At(i)
	}

	return Instantiate(t.Underlying(), args)
}
