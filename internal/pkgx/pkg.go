package pkgx

import (
	"go/types"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"golang.org/x/tools/go/packages"
)

var gPackages sync.Map

const DefaultLoadMode = packages.LoadMode(0b11111111111111111)

func Load(path string) *types.Package {
	_path := path
	if strings.HasSuffix(path, "_test") {
		path = strings.TrimSuffix(_path, "_test")
	}

	pkgs, err := packages.Load(&packages.Config{Tests: true, Mode: DefaultLoadMode}, path)
	msg := "failed to load packages from %s"
	must.NoErrorF(err, msg, path)
	must.BeTrueF(len(pkgs) > 0, msg, path)
	must.BeTrueF(len(pkgs[0].Errors) == 0, msg, path)

	var pkg *types.Package
	for i := range pkgs {
		if pkgs[i].PkgPath == _path {
			pkg = pkgs[i].Types
			break
		}
	}
	must.BeTrue(pkg != nil)
	return pkg
}

func NewT(p *types.Package) Package {
	if p == nil {
		return nil
	}
	return New(p.Path())
}

func New(path string) (p Package) {
	if path == "" {
		return nil
	}

	path = Unwrap(path)

	if v, ok := gPackages.Load(path); ok {
		return v.(Package)
	}

	defer func() {
		gPackages.Store(path, p)
	}()

	return &xpkg{Package: Load(path), id: Wrap(path)}
}

type Package interface {
	Unwrap() *types.Package
	ID() string

	Path() string
	Name() string
	Scope() *types.Scope
}

type xpkg struct {
	id string
	*types.Package
}

func (p *xpkg) Unwrap() *types.Package {
	return p.Package
}

func (p *xpkg) ID() string {
	return p.id
}

func Lookup[T types.Type](p Package, name string) (T, bool) {
	o := p.Scope().Lookup(name)
	if o == nil {
		return *new(T), false
	}
	t, ok := o.Type().(T)
	return t, ok
}

func MustLookup[T types.Type](p Package, name string) T {
	o := p.Scope().Lookup(name)
	if o == nil {
		panic(errors.Errorf("object `%s` not found in %s", name, p.Path()))
	}
	t, ok := o.Type().(T)
	if !ok {
		panic(errors.Errorf("object `%s` is not a %s type", name, reflect.TypeFor[T]()))
	}
	return t
}

func LookupByPath[T types.Type](path, name string) (T, bool) {
	return Lookup[T](New(path), name)
}

func MustLookupByPath[T types.Type](path, name string) T {
	return MustLookup[T](New(path), name)
}
