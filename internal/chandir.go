package internal

import (
	"go/ast"
	"go/types"
	"reflect"

	"github.com/pkg/errors"
)

func NewChanDir(dir any) ChanDir {
	switch d := dir.(type) {
	case types.ChanDir:
		switch d {
		case types.SendOnly:
			return SendDir
		case types.RecvOnly:
			return RecvDir
		case types.SendRecv:
			return BothDir
		}
	case ast.ChanDir:
		switch d {
		case ast.SEND:
			return SendDir
		case ast.RECV:
			return RecvDir
		case ast.ChanDir(3):
			return BothDir
		}
	case reflect.ChanDir:
		switch d {
		case reflect.SendDir:
			return SendDir
		case reflect.RecvDir:
			return RecvDir
		case reflect.BothDir:
			return BothDir
		}
	case ChanDir:
		return d
	}
	panic(errors.Errorf("invalid dir type [%T] %v", dir, dir))
}

type ChanDir reflect.ChanDir

const (
	RecvDir = ChanDir(reflect.RecvDir)
	SendDir = ChanDir(reflect.SendDir)
	BothDir = ChanDir(reflect.BothDir)
)

func (c ChanDir) TypesChanDir() types.ChanDir {
	switch c {
	case SendDir:
		return types.SendOnly
	case RecvDir:
		return types.RecvOnly
	default:
		return types.SendRecv
	}
}

func (c ChanDir) ReflectChanDir() reflect.ChanDir {
	switch c {
	case SendDir:
		return reflect.SendDir
	case RecvDir:
		return reflect.RecvDir
	default:
		return reflect.BothDir
	}
}

func (c ChanDir) AstChanDir() ast.ChanDir {
	switch c {
	case SendDir:
		return ast.SEND
	case RecvDir:
		return ast.RECV
	default:
		return 3
	}
}

func (c ChanDir) String() string {
	switch c {
	case SendDir:
		return "chan<- "
	case RecvDir:
		return "<-chan "
	default:
		return "chan "
	}
}
