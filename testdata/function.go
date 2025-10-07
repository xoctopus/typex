package testdata

import (
	"cmp"
	"fmt"
)

type (
	Func1                           func()
	Func2                           func(string) Int
	Func3                           func(x, y string, z ...int) (Boolean, error)
	Curry                           func(x String, y ...fmt.Stringer) func() string
	Max[T comparable]               func(...T) T
	_Contains[S ~[]E, E comparable] func(s S, v E) bool
	ContainsFunc[S ~[]E, E any]     func(s S, f func(E) bool) bool
	Compare[T cmp.Ordered]          func(x, y T) T
)

type Functions struct {
	Func1              Func1
	Func2              Func2
	Func3              Func3
	Curry              Curry
	Uname              func() func() func() string
	Max                Max[int]
	CompareInt         Compare[int]
	CompareNamedString Compare[String]
}

func (v Max[T]) Compute(e ...T) T {
	return v(e...)
}
