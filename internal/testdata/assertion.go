package testdata

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"

	"github.com/xoctopus/typex/internal"
	"github.com/xoctopus/typex/internal/x"
)

var (
	instantiations [2]func(any) x.Type

	Bundles = []*CompareBundles{
		{name: "Stringer", rtyp: reflect.TypeFor[fmt.Stringer]()},
		{name: "Bytes", rtyp: reflect.TypeFor[interface{ Bytes() []byte }]()},
		{name: "EmptyInterface", rtyp: reflect.TypeFor[any]()},
		{name: "Error", rtyp: reflect.TypeFor[error]()},
		{name: "EmptyStruct", rtyp: reflect.TypeFor[struct{}]()},
		{name: "Struct", rtyp: reflect.TypeFor[struct{ some any }]()},
	}

	Cases = BundleCases{
		&Bundle{name: "Basics", v: Basics{}},
		&Bundle{name: "Composites", v: Composites{}},
		&Bundle{name: "Functions", v: Functions{}},
		&Bundle{name: "Structures", v: Structures{}},
	}
)

func RegisterInstantiations(f1, f2 func(any) x.Type) {
	instantiations[0] = f1
	instantiations[1] = f2

	for _, b := range Bundles {
		b.Init()
	}

	for _, c := range Cases {
		c.Init()
	}
}

type CompareBundle struct {
	name string
	typ  any
}

func (b *CompareBundle) Name() string {
	return b.name
}

func (b *CompareBundle) T() any {
	return b.typ
}

type CompareBundles struct {
	name  string
	rtyp  reflect.Type
	types []*CompareBundle
}

func (b *CompareBundles) Name() string {
	return b.name
}

func (b *CompareBundles) Init() {
	b.types = []*CompareBundle{
		{"ReflectType", b.rtyp},
		{"TypesType", internal.Global().TType(b.rtyp)},
		{"RType", instantiations[0](b.rtyp)},
		{"TType", instantiations[1](b.rtyp)},
	}
}

func (b *CompareBundles) Types() []*CompareBundle {
	return b.types
}

type Bundle struct {
	v     any
	name  string
	cases []*Case
}

func (bc *Bundle) Init() {
	rt := reflect.TypeOf(bc.v)
	must.BeTrue(rt.Kind() == reflect.Struct)

	for i := range rt.NumField() {
		f := rt.Field(i)
		t := f.Type
		name := f.Name
		for range 2 {
			bc.cases = append(bc.cases, &Case{
				name: name,
				r:    t,
				rt:   instantiations[0](t),
				tt:   instantiations[1](t),
			})
			t = reflect.PointerTo(t)
			name = name + "Ptr"
		}
	}
}

func (bc *Bundle) Run(t *testing.T) {
	for _, c := range bc.cases {
		t.Run(c.name, c.Run)
	}
}

func (bc *Bundle) Name() string {
	return bc.name
}

func (bc *Bundle) Cases() []*Case {
	return bc.cases
}

type BundleCases []*Bundle

func NewFieldResult(f x.StructField, exists bool) *FieldResult {
	return &FieldResult{f, exists}
}

type FieldResult struct {
	f      x.StructField
	exists bool
}

type FieldAssertion struct {
	f     reflect.StructField
	exist bool
	typ   string
	xfs   []*FieldResult
}

type FieldAssertions []*FieldAssertion

func (fas FieldAssertions) Run(t *testing.T) {
	for _, fa := range fas {
		for _, r := range fa.xfs {
			if fa.exist {
				NewWithT(t).Expect(r.exists).To(BeTrue())
				NewWithT(t).Expect(r.f).NotTo(BeNil())
				NewWithT(t).Expect(r.f.Name()).To(Equal(fa.f.Name))
				NewWithT(t).Expect(r.f.Tag()).To(Equal(fa.f.Tag))
				NewWithT(t).Expect(r.f.PkgPath()).To(Equal(fa.f.PkgPath))
				NewWithT(t).Expect(r.f.Anonymous()).To(Equal(fa.f.Anonymous))
				NewWithT(t).Expect(r.f.Type().String()).To(Equal(fa.typ))
			} else {
				NewWithT(t).Expect(r.exists).To(BeFalse())
				NewWithT(t).Expect(r.f).To(BeNil())
			}
		}
	}
}

func NewMethodResult(m x.Method, exists bool) *MethodResult {
	return &MethodResult{m, exists}
}

type MethodResult struct {
	m      x.Method
	exists bool
}

type MethodAssertion struct {
	m      reflect.Method
	exists bool
	typ    string
	xms    []*MethodResult
}

type MethodAssertions []*MethodAssertion

func (mas MethodAssertions) Run(t *testing.T) {
	for _, ma := range mas {
		for _, m := range ma.xms {
			if ma.exists {
				NewWithT(t).Expect(m.exists).To(BeTrue())
				NewWithT(t).Expect(m.m).NotTo(BeNil())
				NewWithT(t).Expect(m.m.Name()).To(Equal(ma.m.Name))
				NewWithT(t).Expect(m.m.PkgPath()).To(Equal(ma.m.PkgPath))
				NewWithT(t).Expect(m.m.Type().String()).To(Equal(ma.typ))
			} else {
				NewWithT(t).Expect(m.exists).To(BeFalse())
				NewWithT(t).Expect(m.m).To(BeNil())
			}
		}
	}
}

type Case struct {
	name string
	r    reflect.Type
	rt   x.Type
	tt   x.Type
}

func (c *Case) Name() string {
	return c.name
}

func (c *Case) Run(t *testing.T) {
	t.Run("Kind", func(t *testing.T) {
		NewWithT(t).Expect(c.rt.Kind()).To(Equal(c.r.Kind()))
		NewWithT(t).Expect(c.tt.Kind()).To(Equal(c.r.Kind()))
	})
	t.Run("PkgPath", func(t *testing.T) {
		NewWithT(t).Expect(c.rt.PkgPath()).To(Equal(c.r.PkgPath()))
		NewWithT(t).Expect(c.tt.PkgPath()).To(Equal(c.r.PkgPath()))
	})
	t.Run("Name", func(t *testing.T) {
		NewWithT(t).Expect(c.rt.Name()).To(Equal(c.r.Name()))
		NewWithT(t).Expect(c.tt.Name()).To(Equal(c.r.Name()))
	})
	t.Run("Literal", func(t *testing.T) {
		NewWithT(t).Expect(c.rt.String()).To(Equal(c.tt.String()))
		NewWithT(t).Expect(c.rt.Typename()).To(Equal(c.tt.Typename()))
		NewWithT(t).Expect(c.rt.Alias()).To(Equal(c.tt.Alias()))
	})

	t.Run("Implements", func(t *testing.T) {
		for _, b := range Bundles {
			t.Run(b.Name(), func(t *testing.T) {
				expect := b.rtyp.Kind() == reflect.Interface && c.r.Implements(b.rtyp)
				for _, v := range b.Types() {
					t.Run(v.Name(), func(t *testing.T) {
						NewWithT(t).Expect(c.rt.Implements(v.T())).To(Equal(expect))
						NewWithT(t).Expect(c.tt.Implements(v.T())).To(Equal(expect))
					})
				}
			})
		}
		t.Run("Nil", func(t *testing.T) {
			NewWithT(t).Expect(c.rt.Implements(nil)).To(BeFalse())
			NewWithT(t).Expect(c.tt.Implements(nil)).To(BeFalse())
		})
	})

	t.Run("AssignableTo", func(t *testing.T) {
		for _, b := range Bundles {
			t.Run(b.Name(), func(t *testing.T) {
				expect := c.r.AssignableTo(b.rtyp)
				for _, v := range b.Types() {
					t.Run(v.Name(), func(t *testing.T) {
						NewWithT(t).Expect(c.rt.AssignableTo(v.T())).To(Equal(expect))
						NewWithT(t).Expect(c.tt.AssignableTo(v.T())).To(Equal(expect))
					})
				}
			})
		}
		t.Run("Nil", func(t *testing.T) {
			NewWithT(t).Expect(c.rt.AssignableTo(nil)).To(BeFalse())
			NewWithT(t).Expect(c.tt.AssignableTo(nil)).To(BeFalse())
		})
	})

	t.Run("ConvertibleTo", func(t *testing.T) {
		for _, b := range Bundles {
			t.Run(b.Name(), func(t *testing.T) {
				expect := c.r.ConvertibleTo(b.rtyp)
				for _, v := range b.Types() {
					t.Run(v.Name(), func(t *testing.T) {
						NewWithT(t).Expect(c.rt.ConvertibleTo(v.T())).To(Equal(expect))
						NewWithT(t).Expect(c.tt.ConvertibleTo(v.T())).To(Equal(expect))
					})
				}
			})
		}
		t.Run("Nil", func(t *testing.T) {
			NewWithT(t).Expect(c.rt.ConvertibleTo(nil)).To(BeFalse())
			NewWithT(t).Expect(c.tt.ConvertibleTo(nil)).To(BeFalse())
		})
	})

	t.Run("Comparable", func(t *testing.T) {
		expect := c.r.Comparable()
		NewWithT(t).Expect(c.rt.Comparable()).To(Equal(expect))
		NewWithT(t).Expect(c.tt.Comparable()).To(Equal(expect))
	})

	t.Run("Key", func(t *testing.T) {
		if c.r.Kind() == reflect.Map {
			NewWithT(t).Expect(c.rt.Key().String()).To(Equal(c.tt.Key().String()))
		} else {
			NewWithT(t).Expect(c.rt.Key()).To(BeNil())
			NewWithT(t).Expect(c.tt.Key()).To(BeNil())
		}
	})

	t.Run("Elem", func(t *testing.T) {
		if reflectx.CanElem(c.r.Kind()) {
			NewWithT(t).Expect(c.rt.Elem().String()).To(Equal(c.tt.Elem().String()))
		} else {
			NewWithT(t).Expect(c.rt.Elem()).To(BeNil())
			NewWithT(t).Expect(c.tt.Elem()).To(BeNil())
		}
	})

	t.Run("Len", func(t *testing.T) {
		if c.r.Kind() == reflect.Array {
			NewWithT(t).Expect(c.rt.Len()).To(Equal(c.r.Len()))
			NewWithT(t).Expect(c.tt.Len()).To(Equal(c.r.Len()))
		} else {
			NewWithT(t).Expect(c.rt.Len()).To(Equal(0))
			NewWithT(t).Expect(c.tt.Len()).To(Equal(0))
		}
	})

	fields := 0
	if c.r.Kind() == reflect.Struct {
		fields = c.r.NumField()
	}

	t.Run("NumField", func(t *testing.T) {
		NewWithT(t).Expect(c.rt.NumField()).To(Equal(fields))
		NewWithT(t).Expect(c.tt.NumField()).To(Equal(fields))
	})

	t.Run("Field", func(t *testing.T) {
		fas := c.FieldAssertions(fields, "Name", "name", "str", "_", "unexported", "")
		fas.Run(t)
		t.Run("OutOfRange", func(t *testing.T) {
			NewWithT(t).Expect(c.rt.Field(fields + 1)).To(BeNil())
			NewWithT(t).Expect(c.tt.Field(fields + 1)).To(BeNil())
		})
		t.Run("MatchAlwaysTrue", func(t *testing.T) {
			match := func(v string) bool { return true }
			rf, ok1 := c.rt.FieldByNameFunc(match)
			tf, ok2 := c.tt.FieldByNameFunc(match)
			if fields != 1 {
				NewWithT(t).Expect(rf).To(BeNil())
				NewWithT(t).Expect(ok1).To(BeFalse())
				NewWithT(t).Expect(tf).To(BeNil())
				NewWithT(t).Expect(ok2).To(BeFalse())
			} else {
				rf0 := c.rt.Field(0)
				NewWithT(t).Expect(ok1).To(BeTrue())
				NewWithT(t).Expect(rf.Type().String()).To(Equal(rf0.Type().String()))
				tf0 := c.tt.Field(0)
				NewWithT(t).Expect(ok2).To(BeTrue())
				NewWithT(t).Expect(tf.Type().String()).To(Equal(tf0.Type().String()))
			}
		})
	})

	methods := c.r.NumMethod()
	t.Run("NumMethod", func(t *testing.T) {
		NewWithT(t).Expect(c.rt.NumMethod()).To(Equal(methods))
		NewWithT(t).Expect(c.tt.NumMethod()).To(Equal(methods))
	})

	t.Run("Method", func(t *testing.T) {
		mas := c.MethodAssertions(methods, "Name", "name", "str", "_", "unexported", "")
		mas.Run(t)
		t.Run("OutOfRange", func(t *testing.T) {
			NewWithT(t).Expect(c.rt.Method(methods + 1)).To(BeNil())
			NewWithT(t).Expect(c.tt.Method(methods + 1)).To(BeNil())
		})
	})

	t.Run("IsVariadic", func(t *testing.T) {
		expect := false
		if c.r.Kind() == reflect.Func {
			expect = c.r.IsVariadic()
		}
		NewWithT(t).Expect(c.rt.IsVariadic()).To(Equal(expect))
		NewWithT(t).Expect(c.tt.IsVariadic()).To(Equal(expect))
	})

	t.Run("Ins", func(t *testing.T) {
		if c.r.Kind() == reflect.Func {
			NewWithT(t).Expect(c.rt.NumIn()).To(Equal(c.r.NumIn()))
			NewWithT(t).Expect(c.tt.NumIn()).To(Equal(c.r.NumIn()))
			for i := range c.r.NumIn() {
				NewWithT(t).Expect(c.rt.In(i).Name()).To(Equal(c.tt.In(i).Name()))
				NewWithT(t).Expect(c.rt.In(i).String()).To(Equal(c.tt.In(i).String()))
				NewWithT(t).Expect(c.rt.In(i).PkgPath()).To(Equal(c.tt.In(i).PkgPath()))
			}
			t.Run("OutOfRange", func(t *testing.T) {
				NewWithT(t).Expect(c.rt.In(c.r.NumIn())).To(BeNil())
				NewWithT(t).Expect(c.tt.In(c.r.NumIn())).To(BeNil())
			})
		} else {
			NewWithT(t).Expect(c.rt.NumIn()).To(Equal(0))
			NewWithT(t).Expect(c.tt.NumIn()).To(Equal(0))
			NewWithT(t).Expect(c.rt.In(0)).To(BeNil())
			NewWithT(t).Expect(c.tt.In(0)).To(BeNil())
		}
	})

	t.Run("Outs", func(t *testing.T) {
		if c.r.Kind() == reflect.Func {
			NewWithT(t).Expect(c.rt.NumOut()).To(Equal(c.r.NumOut()))
			NewWithT(t).Expect(c.tt.NumOut()).To(Equal(c.r.NumOut()))
			for i := range c.r.NumOut() {
				NewWithT(t).Expect(c.rt.Out(i).Name()).To(Equal(c.tt.Out(i).Name()))
				NewWithT(t).Expect(c.rt.Out(i).String()).To(Equal(c.tt.Out(i).String()))
				NewWithT(t).Expect(c.rt.Out(i).PkgPath()).To(Equal(c.tt.Out(i).PkgPath()))
			}
			t.Run("OutOfRange", func(t *testing.T) {
				NewWithT(t).Expect(c.rt.Out(c.r.NumOut())).To(BeNil())
				NewWithT(t).Expect(c.tt.Out(c.r.NumOut())).To(BeNil())
			})
		} else {
			NewWithT(t).Expect(c.rt.NumOut()).To(Equal(0))
			NewWithT(t).Expect(c.tt.NumOut()).To(Equal(0))
			NewWithT(t).Expect(c.rt.Out(0)).To(BeNil())
			NewWithT(t).Expect(c.tt.Out(0)).To(BeNil())
		}
	})
}

func (c *Case) FieldAssertions(n int, options ...string) FieldAssertions {
	var fas FieldAssertions
	for i := range n {
		f := c.r.Field(i)
		match := func(v string) bool { return v == f.Name }

		rfi := c.rt.Field(i)
		tfi := c.tt.Field(i)

		fa := &FieldAssertion{
			f, true, rfi.Type().String(),
			[]*FieldResult{
				NewFieldResult(rfi, rfi != nil),
				NewFieldResult(tfi, tfi != nil),
				NewFieldResult(c.rt.FieldByName(f.Name)),
				NewFieldResult(c.tt.FieldByName(f.Name)),
				NewFieldResult(c.rt.FieldByNameFunc(match)),
				NewFieldResult(c.tt.FieldByNameFunc(match)),
			},
		}

		fas = append(fas, fa)
	}
	for _, name := range options {
		match := func(v string) bool { return v == name }

		var (
			f      reflect.StructField
			exists bool
		)
		if c.r.Kind() == reflect.Struct {
			f, exists = c.r.FieldByName(name)
		}
		fa := &FieldAssertion{
			f, exists, "",
			[]*FieldResult{
				NewFieldResult(c.rt.FieldByName(name)),
				NewFieldResult(c.tt.FieldByName(name)),
				NewFieldResult(c.rt.FieldByNameFunc(match)),
				NewFieldResult(c.tt.FieldByNameFunc(match)),
			},
		}
		if exists {
			fa.typ = fa.xfs[0].f.Type().String()
		}
		fas = append(fas, fa)
	}
	return fas
}

func (c *Case) MethodAssertions(n int, options ...string) MethodAssertions {
	var mas MethodAssertions
	for i := range n {
		m := c.r.Method(i)

		rmi := c.rt.Method(i)
		tmi := c.tt.Method(i)

		mas = append(mas, &MethodAssertion{
			m: m, exists: true, typ: rmi.Type().String(),
			xms: []*MethodResult{
				NewMethodResult(rmi, rmi != nil),
				NewMethodResult(tmi, tmi != nil),
				NewMethodResult(c.rt.MethodByName(m.Name)),
				NewMethodResult(c.tt.MethodByName(m.Name)),
			},
		})
	}

	for _, name := range options {
		m, exists := c.r.MethodByName(name)
		ma := &MethodAssertion{
			m: m, exists: exists, typ: "",
			xms: []*MethodResult{
				NewMethodResult(c.rt.MethodByName(name)),
				NewMethodResult(c.tt.MethodByName(name)),
			},
		}
		if xm := ma.xms[0]; xm.m != nil {
			ma.typ = xm.m.Type().String()
		}
		mas = append(mas, ma)
	}
	return mas
}
