package testdata

import (
	"net"
)

type Generics[T any] struct {
	AliasInt  AliasInt
	Array     [1]T
	Basic     int
	Chan      chan T
	ChanW     <-chan T
	ChanR     chan<- T
	Map       map[string]T
	Pointer   *T
	Slice     []T
	Struct    struct{ FuncT func() T }
	Interface interface{ Value() T }
	TypedArray[T]
	TypedSlice[net.Addr]
	NoTArg Int
}
