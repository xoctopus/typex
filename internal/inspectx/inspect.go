package inspectx

import (
	"go/ast"
	"go/types"
	"slices"
	"sort"

	"github.com/xoctopus/x/reflectx"
)

type Field struct {
	f   *types.Var
	tag string
}

func (f *Field) Var() *types.Var {
	return f.f
}

func (f *Field) Tag() string {
	return f.tag
}

type Fields map[string][]*Field

type Method struct {
	f     *types.Func
	refs  []types.Type
	level int
}

// CheckRef checks if this method can reference by t. when a method's receiver
// is a pointer type. it needs a pointer type to refer it in the derived list of
// this method.
func (m *Method) CheckRef(t types.Type) bool {
	recv := m.f.Signature().Recv().Type()
	if _, ok := recv.(*types.Pointer); !ok {
		return true
	}
	if _, ok := t.(*types.Pointer); ok {
		return true
	}
	for i := len(m.refs) - 1; i >= 0; i-- {
		if _, ok := m.refs[i].(*types.Pointer); ok {
			return true
		}
	}
	return false
}

type Methods map[string][]*Method

func InspectMethods(t types.Type) []*types.Func {
	i := &inspector{t: t, fields: make(Fields), methods: make(Methods)}
	i.inspect(t)

	return i.unambiguous()
}

type inspector struct {
	t       types.Type
	fields  Fields
	methods Methods
	walker  Walker
}

func (i *inspector) appendField(v *types.Var, tag string) {
	i.fields[v.Name()] = append(i.fields[v.Name()], &Field{v, tag})
}

func (i *inspector) appendMethod(f *types.Func) {
	if !ast.IsExported(f.Name()) {
		return
	}

	method := &Method{f, slices.Clone(i.walker), 0}
	for _, ref := range i.walker {
		if _, ok := ref.(*types.Named); ok {
			method.level++
		}
	}

	i.methods[f.Name()] = append(i.methods[f.Name()], method)
}

func (i *inspector) inspect(t types.Type) {
	t = types.Unalias(t)

	if i.walker.IsVisited(t) {
		return
	}
	defer func() {
		i.walker.Visited(t)
	}()
	i.walker.Visit(t)

	switch x := t.(type) {
	case *types.Pointer:
		e := x.Elem()
		if reflectx.CanCast[*types.Pointer](e) {
			return
		}
		if reflectx.CanCast[*types.Interface](e.Underlying()) {
			return
		}
		i.inspect(e)
	case *types.Named:
		i.inspect(x.Underlying())
		for idx := range x.NumMethods() {
			i.appendMethod(x.Method(idx))
		}
	case *types.Interface:
		for idx := range x.NumMethods() {
			i.appendMethod(x.Method(idx))
		}
	case *types.Struct:
		for idx := range x.NumFields() {
			f := x.Field(idx)
			i.appendField(f, x.Tag(idx))
			if f.Anonymous() {
				i.inspect(f.Type())
			}
		}
	}
	return
}

func (i *inspector) unambiguous() []*types.Func {
	final := make([]*types.Func, 0, len(i.methods))
	for name, methods := range i.methods {
		if _, ok := i.fields[name]; ok {
			continue
		}
		if len(methods) == 1 {
			if methods[0].CheckRef(i.t) {
				final = append(final, methods[0].f)
			}
			continue
		}
		sort.Slice(methods, func(i, j int) bool {
			return methods[i].level < methods[j].level
		})
		if methods[0].level < methods[1].level && methods[0].CheckRef(i.t) {
			final = append(final, methods[0].f)
		}
	}
	sort.Slice(final, func(i, j int) bool {
		return final[i].Name() < final[j].Name()
	})
	return final
}

func FieldByNameFunc(t types.Type, match func(string) bool) *Field {
	return InspectField(t, match, Walker{}, 0)
}

func FieldByName(t types.Type, name string) *Field {
	return FieldByNameFunc(t, func(s string) bool { return s == name })
}

func InspectField(t types.Type, match func(string) bool, w Walker, entries int) *Field {
	t = types.Unalias(t)

	if w.IsVisited(t) {
		return nil
	}
	defer func() {
		w.Visited(t)
	}()
	w.Visit(t)

	switch x := t.(type) {
	case *types.Named:
		return InspectField(t.Underlying(), match, w, entries)
	case *types.Pointer:
		if entries == 0 {
			return nil
		}
		return InspectField(x.Elem(), match, w, entries)
	case *types.Struct:
		var (
			directed  *Field
			embeddeds []*Field
		)
		for i := range x.NumFields() {
			f := x.Field(i)
			if match(f.Name()) {
				if directed != nil {
					return nil
				}
				directed = &Field{f, x.Tag(i)}
			}
			if f.Anonymous() {
				if field := InspectField(f.Type(), match, w, entries+1); field != nil {
					embeddeds = append(embeddeds, field)
				}
			}
		}
		if directed != nil {
			return directed
		}
		if len(embeddeds) == 1 {
			return embeddeds[0]
		}
		return nil
	default:
		return nil
	}
}
