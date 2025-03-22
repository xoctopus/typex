package testdata

import (
	"fmt"
	"net"

	"github.com/xoctopus/x/reflectx"
)

type Tagged struct {
	A              string       `json:"a"`
	B              String       `json:"b"`
	Namer          fmt.Stringer `json:"-,\"'#{}[]()<>!@#$%^&*_-+=\\|\""`
	EmptyInterface `json:"-"`   // anonymous
	unexported     any
}

type Embedded struct {
	Tagged
	Name string
}

type Uncomparable struct {
	v map[any]any
}

type Serialized[T CanBeSerialized] struct {
	data T
}

func (v Serialized[T]) String() string {
	switch data := any(v.data).(type) {
	case []byte:
		return string(data)
	default:
		return reflectx.MustAssertType[string](data)
	}
}

func (v Serialized[T]) Bytes() []byte {
	switch data := any(v.data).(type) {
	case string:
		return []byte(data)
	default:
		return reflectx.MustAssertType[[]byte](data)
	}
}

func (v Serialized[T]) Data() T { return v.data }

func (v *Serialized[T]) SetData(data T) { v.data = data }

type BTreeNode[T any] struct {
	v       T
	r, l, p *BTreeNode[T]
}

func (n *BTreeNode[T]) InsertL(v T) *BTreeNode[T] {
	l := &BTreeNode[T]{v: v, p: n}
	orphan := n.l
	if orphan != nil {
		orphan.p = nil
	}
	n.l = l
	return orphan
}

func (n *BTreeNode[T]) InsertR(v T) *BTreeNode[T] {
	r := &BTreeNode[T]{v: v, p: n}
	orphan := n.r
	if orphan != nil {
		orphan.p = nil
	}
	n.r = r
	return orphan
}

type PassTypeParam[T1 any, T2 fmt.Stringer] struct {
	v1 T1
	v2 T2
	*BTreeNode[T2]
}

func (v *PassTypeParam[T1, T2]) Deal(v1 T1) T2 {
	v.v1 = v1
	return v.v2
}

type CircleEmbedsA struct {
	CircleEmbedsB
}

type CircleEmbedsB struct {
	*CircleEmbedsC
}

type CircleEmbedsC struct {
	CircleEmbedsA
	*CircleEmbedsB
}

type HasUnexportedMethod struct{}

func (HasUnexportedMethod) str() string { return "HasUnexportedMethod" }

// StringerL1 strict call path `StringerL1.String`
type StringerL1 string

func (v *StringerL1) String() string { return string(*v) }

// StringerL2 strict call path `StringerL1.Stringer.String`
type StringerL2 struct {
	fmt.Stringer
}

// StringerL2WrapL1 strict call path `StringerL2WrapL1.StringerL1.String`
type StringerL2WrapL1 struct {
	*StringerL1
}

// StringerL3 strict call path `StringerL3.StringerL2.Stringer.String`
type StringerL3 struct {
	StringerL2
}

// StringerL3WrapL2 strict call path `StringerL3WrapL2.StringerL2WrapL1.StringerL1.String`
type StringerL3WrapL2 struct {
	StringerL2WrapL1
}

type StringField struct {
	String string
}

type AmbiguousL1x2 struct {
	StringerL1
	fmt.Stringer
}

type AmbiguousL1AndField struct {
	StringerL1
	String string
}

type AmbiguousL2x2 struct {
	StringerL2
	*StringerL2WrapL1
}

type UnambiguousL1AndL2x2 struct {
	*StringerL2
	StringerL2WrapL1
	*StringerL1
}

type AmbiguousL1x2AndL2 struct {
	StringerL1
	fmt.Stringer
	StringerL2
}

type UnambiguousL2AndL3x2 struct {
	StringerL2
	StringerL3
	StringerL3WrapL2
}

type AmbiguousL2AndL3x2AndField struct {
	StringerL2
	StringerL3
	StringerL3WrapL2
	StringField
}

type Structures struct {
	Tagged                     Tagged
	Embedded                   Embedded
	Uncomparable               Uncomparable
	SerializedString           Serialized[string]
	SerializedBytes            Serialized[[]byte]
	BTreeNodeInt               BTreeNode[int]
	BTreeNodeTypedString       BTreeNode[string]
	PassTypeParam1             PassTypeParam[int, net.Addr]
	PassTypeParam2             PassTypeParam[int, Serialized[string]]
	PassTypeParam3             PassTypeParam[int, Serialized[[]byte]]
	HasUnexportedMethod        HasUnexportedMethod
	CircleEmbedsA              CircleEmbedsA
	CircleEmbedsB              CircleEmbedsB
	CircleEmbedsC              CircleEmbedsC
	StringerL1                 StringerL1
	StringerL2                 StringerL2
	StringerL2WrapL1           StringerL2WrapL1
	StringerL3                 StringerL3
	StringerL3WrapL2           StringerL3WrapL2
	StringField                StringField
	AmbiguousL1x2              AmbiguousL1x2
	AmbiguousL1AndField        AmbiguousL1AndField
	AmbiguousL2x2              AmbiguousL2x2
	UnambiguousL1AndL2x2       UnambiguousL1AndL2x2
	AmbiguousL1x2AndL2         AmbiguousL1x2AndL2
	UnambiguousL2AndL3x2       UnambiguousL2AndL3x2
	AmbiguousL2AndL3x2AndField AmbiguousL2AndL3x2AndField
}
