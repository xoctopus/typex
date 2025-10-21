package internal

import (
	"context"
	"go/types"
	"reflect"
	"unsafe"

	"github.com/xoctopus/typex/pkgutil"
)

var builtins = []*builtin{
	{
		kind:     reflect.Bool,
		typename: "bool",
		rtyp:     reflect.TypeFor[bool](),
		ttyp:     types.Typ[types.Bool],
	}, {
		kind:     reflect.Int,
		typename: "int",
		rtyp:     reflect.TypeFor[int](),
		ttyp:     types.Typ[types.Int],
	}, {
		kind:     reflect.Int8,
		typename: "int8",
		rtyp:     reflect.TypeFor[int8](),
		ttyp:     types.Typ[types.Int8],
	}, {
		kind:     reflect.Int16,
		typename: "int16",
		rtyp:     reflect.TypeFor[int16](),
		ttyp:     types.Typ[types.Int16],
	}, {
		kind:     reflect.Int32,
		typename: "int32",
		alias:    "rune",
		rtyp:     reflect.TypeFor[int32](),
		ttyp:     types.Typ[types.Int32],
	}, {
		kind:     reflect.Int64,
		typename: "int64",
		rtyp:     reflect.TypeFor[int64](),
		ttyp:     types.Typ[types.Int64],
	}, {
		kind:     reflect.Uint,
		typename: "uint",
		rtyp:     reflect.TypeFor[uint](),
		ttyp:     types.Typ[types.Uint],
	}, {
		kind:     reflect.Uint8,
		typename: "uint8",
		rtyp:     reflect.TypeFor[uint8](),
		alias:    "byte",
		ttyp:     types.Typ[types.Uint8],
	}, {
		kind:     reflect.Uint16,
		typename: "uint16",
		rtyp:     reflect.TypeFor[uint16](),
		ttyp:     types.Typ[types.Uint16],
	}, {
		kind:     reflect.Uint32,
		typename: "uint32",
		rtyp:     reflect.TypeFor[uint32](),
		ttyp:     types.Typ[types.Uint32],
	}, {
		kind:     reflect.Uint64,
		typename: "uint64",
		rtyp:     reflect.TypeFor[uint64](),
		ttyp:     types.Typ[types.Uint64],
	}, {
		kind:     reflect.Uintptr,
		typename: "uintptr",
		rtyp:     reflect.TypeFor[uintptr](),
		ttyp:     types.Typ[types.Uintptr],
	}, {
		kind:     reflect.Float32,
		typename: "float32",
		rtyp:     reflect.TypeFor[float32](),
		ttyp:     types.Typ[types.Float32],
	}, {
		kind:     reflect.Float64,
		typename: "float64",
		rtyp:     reflect.TypeFor[float64](),
		ttyp:     types.Typ[types.Float64],
	}, {
		kind:     reflect.Complex64,
		typename: "complex64",
		rtyp:     reflect.TypeFor[complex64](),
		ttyp:     types.Typ[types.Complex64],
	}, {
		kind:     reflect.Complex128,
		typename: "complex128",
		rtyp:     reflect.TypeFor[complex128](),
		ttyp:     types.Typ[types.Complex128],
	}, {
		kind:     reflect.String,
		typename: "string",
		rtyp:     reflect.TypeFor[string](),
		ttyp:     types.Typ[types.String],
	}, {
		kind:     reflect.UnsafePointer,
		pkg:      pkgutil.New("unsafe"),
		typename: "Pointer",
		rtyp:     reflect.TypeFor[unsafe.Pointer](),
		ttyp:     types.Typ[types.UnsafePointer],
	}, {
		kind:     reflect.Interface,
		typename: "interface {}",
		alias:    "any",
		rtyp:     reflect.TypeFor[any](),
		ttyp:     types.NewInterfaceType(nil, nil),
	}, {
		kind:     reflect.Interface,
		typename: "error",
		rtyp:     reflect.TypeFor[error](),
		ttyp:     pkgutil.MustLookupByPath[*types.Signature]("errors", "New").Results().At(0).Type(),
	},
}

type builtin struct {
	kind     reflect.Kind
	pkg      pkgutil.Package
	typename string
	alias    string
	rtyp     reflect.Type
	ttyp     types.Type
}

type Builtin interface {
	Literal
	Kind() reflect.Kind
	Alias() string
}

var _ Builtin = (*builtin)(nil)

func (t *builtin) Kind() reflect.Kind {
	return t.kind
}

func (t *builtin) PkgPath() string {
	if t.pkg != nil {
		return t.pkg.Path()
	}
	return ""
}

func (t *builtin) Name() string {
	if t.typename == "interface {}" {
		return ""
	}
	return t.typename
}

func (t *builtin) String() string {
	if t.pkg != nil {
		return t.pkg.Path() + "." + t.typename
	}
	return t.typename
}

func (t *builtin) TypeLit(_ context.Context) string {
	return t.String()
}

func (t *builtin) Alias() string {
	return t.alias
}

func (t *builtin) TType() types.Type {
	return t.ttyp
}
