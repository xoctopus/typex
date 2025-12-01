package typx

import (
	"errors"
	"go/types"
	"reflect"
	"slices"
	"strings"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/syncx"
	gopkg "golang.org/x/tools/go/packages"
)

var gPackages = syncx.NewXmap[string, *types.Package]()

func Load(path string) (p *types.Package) {
	if x, ok := gPackages.Load(path); ok {
		return x
	}

	defer func() {
		if p != nil {
			gPackages.Store(p.Path(), p)
		}
	}()

	_path := path
	if strings.HasSuffix(path, "_test") {
		path = strings.TrimSuffix(_path, "_test")
	}

	pkgs, err := gopkg.Load(&gopkg.Config{Mode: 9183, Tests: true}, path)
	must.NoErrorF(err, "failed to load %s", path)
	must.BeTrueF(len(pkgs) > 0, "no packages loaded")
	must.NoErrorF(errors.Join(
		slices.Collect(func(yield func(error) bool) {
			for _, pkg := range pkgs {
				for _, x := range pkg.Errors {
					yield(x)
				}
			}
		})...,
	), "failed to load %s", path)

	for i := range pkgs {
		if pkgs[i].PkgPath == _path {
			p = pkgs[i].Types
			break
		}
	}
	must.BeTrueF(p != nil, "failed to load %s", path)
	return p
}

func Lookup[T types.Type](p *types.Package, name string) T {
	obj := p.Scope().Lookup(name)
	must.BeTrueF(
		obj != nil,
		"must lookup %s.%s to %s",
		p.Path(), name, reflect.TypeFor[T](),
	)

	typ, ok := obj.Type().(T)
	must.BeTrueF(
		ok,
		"must lookup %s.%s to %s",
		p.Path(), name, reflect.TypeFor[T](),
	)
	return typ
}
