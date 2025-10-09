package internal_test

import (
	"context"
	"fmt"
	"go/types"
	"io"
	"iter"
	"net"
	"reflect"
	"testing"
	"unsafe"

	"github.com/xoctopus/x/resultx"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typex/internal"
	"github.com/xoctopus/typex/namer"
	"github.com/xoctopus/typex/pkgutil"
	"github.com/xoctopus/typex/testdata"
)

var (
	g = internal.Global()
	p = pkgutil.New("github.com/xoctopus/typex/testdata")

	tTestdataTagged                 = pkgutil.MustLookup[*types.Named](p, "Tagged")
	tTestdataTypedSliceAliasNetAddr = pkgutil.MustLookup[*types.Alias](p, "TypedSliceAliasNetAddr")
	tTestdataMap                    = pkgutil.MustLookup[*types.Named](p, "Map")
	tTestdataPassTypeParam          = pkgutil.MustLookup[*types.Named](p, "PassTypeParam")
	tTestdataTypedArray             = pkgutil.MustLookup[*types.Named](p, "TypedArray")
	tTestdataTypedSlice             = pkgutil.MustLookup[*types.Named](p, "TypedSlice")

	tFmtStringer   = pkgutil.MustLookupByPath[*types.Named]("fmt", "Stringer")
	tIoReadCloser  = pkgutil.MustLookupByPath[*types.Named]("io", "ReadCloser")
	tIoWriteCloser = pkgutil.MustLookupByPath[*types.Named]("io", "WriteCloser")
	tIoReadWriter  = pkgutil.MustLookupByPath[*types.Named]("io", "ReadWriter")

	tNetAddr        = pkgutil.MustLookupByPath[*types.Named]("net", "Addr")
	tError          = pkgutil.MustLookupByPath[*types.Signature]("errors", "New").Results().At(0).Type()
	tEmptyInterface = types.NewInterfaceType(nil, nil)
	tEmptyStruct    = types.NewStruct(nil, nil)
)

var GlobalCases = []struct {
	rt      reflect.Type
	tt      types.Type
	id      string
	wrapped string
	pkg     string
	name    string
	literal string
}{
	{reflect.Type(nil), types.Type(nil), "", "", "", "", ""},
	{
		reflect.TypeFor[bool](),
		types.Typ[types.Bool],
		"bool", "bool", "", "bool", "bool",
	},
	{
		reflect.TypeFor[unsafe.Pointer](),
		types.Typ[types.UnsafePointer],
		"unsafe.Pointer", "unsafe.Pointer", "unsafe", "Pointer", "unsafe.Pointer",
	},
	{
		reflect.TypeFor[any](),
		types.NewInterfaceType(nil, nil),
		"interface {}", "interface {}", "", "", "interface {}",
	},
	{
		reflect.TypeFor[rune](),
		types.Typ[types.Rune],
		"int32", "int32", "", "int32", "int32",
	},
	{
		reflect.TypeFor[byte](),
		types.Typ[types.Byte],
		"uint8", "uint8", "", "uint8", "uint8",
	},
	{
		reflect.TypeFor[[]int](),
		types.NewSlice(types.Typ[types.Int]),
		"[]int", "[]int", "", "", "[]int",
	},
	{
		reflect.TypeFor[[]byte](),
		types.NewSlice(types.Typ[types.Byte]),
		"[]uint8", "[]uint8", "", "", "[]uint8",
	},
	{
		reflect.TypeFor[[3]rune](),
		types.NewArray(types.Typ[types.Rune], 3),
		"[3]int32", "[3]int32", "", "", "[3]int32",
	},
	{
		reflect.TypeFor[[3]iter.Seq[int]](),
		types.NewArray(resultx.Unwrap(types.Instantiate(
			nil, pkgutil.MustLookup[*types.Named](pkgutil.New("iter"), "Seq"),
			[]types.Type{types.Typ[types.Int]}, true,
		)), 3),
		"[3]iter.Seq[int]", "[3]iter.Seq[int]", "", "", "[3]iter.Seq[int]",
	},
	{
		reflect.TypeFor[chan error](),
		types.NewChan(types.SendRecv, g.TType(reflect.TypeFor[error]())),
		"chan error", "chan error", "", "", "chan error",
	},
	{
		reflect.TypeFor[chan<- testdata.Tagged](),
		types.NewChan(types.SendOnly, tTestdataTagged),
		"chan<- github.com/xoctopus/typex/testdata.Tagged",
		"chan<- xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.Tagged",
		"", "",
		"chan<- testdata.Tagged",
	},
	{
		reflect.TypeFor[<-chan *testdata.Tagged](),
		types.NewChan(types.RecvOnly, types.NewPointer(tTestdataTagged)),
		"<-chan *github.com/xoctopus/typex/testdata.Tagged",
		"<-chan *xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.Tagged",
		"", "",
		"<-chan *testdata.Tagged",
	},
	{
		reflect.TypeFor[[]testdata.Tagged](),
		types.NewSlice(tTestdataTagged),
		"[]github.com/xoctopus/typex/testdata.Tagged",
		"[]xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.Tagged",
		"", "",
		"[]testdata.Tagged",
	},
	{
		reflect.TypeFor[func()](),
		types.NewSignatureType(nil, nil, nil, nil, nil, false),
		"func()", "func()", "", "", "func()",
	},
	{
		reflect.TypeFor[func(fmt.Stringer, ...any) net.Addr](),
		types.NewSignatureType(
			nil, nil, nil,
			types.NewTuple(
				types.NewParam(0, nil, "", tFmtStringer),
				types.NewParam(0, nil, "", types.NewSlice(tEmptyInterface)),
			),
			types.NewTuple(types.NewParam(0, nil, "", tNetAddr)),
			true,
		),
		"func(fmt.Stringer, ...interface {}) net.Addr",
		"func(fmt.Stringer, ...interface {}) net.Addr",
		"", "",
		"func(fmt.Stringer, ...interface {}) net.Addr",
	},
	{
		reflect.TypeFor[func(string, ...any) (string, error)](),
		types.NewSignatureType(
			nil, nil, nil,
			types.NewTuple(
				types.NewParam(0, nil, "", types.Typ[types.String]),
				types.NewParam(0, nil, "", types.NewSlice(tEmptyInterface)),
			),
			types.NewTuple(
				types.NewParam(0, nil, "", types.Typ[types.String]),
				types.NewParam(0, nil, "", tError),
			),
			true,
		),
		"func(string, ...interface {}) (string, error)",
		"func(string, ...interface {}) (string, error)",
		"", "",
		"func(string, ...interface {}) (string, error)",
	},
	{
		reflect.TypeFor[interface {
			fmt.Stringer
			io.ReadCloser
			io.WriteCloser
			io.ReadWriter
		}](),
		types.NewInterfaceType(nil, []types.Type{tFmtStringer, tIoReadCloser, tIoWriteCloser, tIoReadWriter}),
		"interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }",
		"interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }",
		"", "",
		"interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }",
	},
	{
		reflect.TypeFor[testdata.TypedSliceAliasNetAddr](),
		tTestdataTypedSliceAliasNetAddr,
		"github.com/xoctopus/typex/testdata.TypedSlice[net.Addr]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedSlice[net.Addr]",
		"github.com/xoctopus/typex/testdata",
		"TypedSlice[net.Addr]",
		"testdata.TypedSlice[net.Addr]",
	},
	{
		reflect.TypeFor[struct{}](),
		tEmptyStruct,
		"struct {}", "struct {}", "", "", "struct {}",
	},
	{
		reflect.TypeFor[struct {
			A            string
			B            int
			testdata.Map `json:"esc''{}[]\""`
		}](),
		types.NewStruct([]*types.Var{
			types.NewField(0, nil, "A", types.Typ[types.String], false),
			types.NewField(0, nil, "B", types.Typ[types.Int], false),
			types.NewField(0, p.Unwrap(), "", tTestdataMap, true),
		}, []string{"", "", `json:"esc''{}[]\""`}),
		`struct { A string; B int; github.com/xoctopus/typex/testdata.Map "json:\"esc''{}[]\\\"\"" }`,
		`struct { A string; B int; xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.Map "json:\"esc''{}[]\\\"\"" }`,
		"", "",
		`struct { A string; B int; testdata.Map "json:\"esc''{}[]\\\"\"" }`,
	},
	{
		reflect.TypeFor[map[string]int](),
		types.NewMap(types.Typ[types.String], types.Typ[types.Int]),
		"map[string]int", "map[string]int", "", "", "map[string]int",
	},
	{
		reflect.TypeFor[testdata.PassTypeParam[int, net.Addr]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataPassTypeParam, []types.Type{
				types.Typ[types.Int], tNetAddr,
			}, true),
		),
		"github.com/xoctopus/typex/testdata.PassTypeParam[int,net.Addr]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.PassTypeParam[int,net.Addr]",
		"github.com/xoctopus/typex/testdata",
		"PassTypeParam[int,net.Addr]",
		"testdata.PassTypeParam[int,net.Addr]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[[]string]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewSlice(types.Typ[types.String])}, true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[[]string]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[[]string]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[[]string]",
		"testdata.TypedArray[[]string]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[[2]string]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewArray(types.Typ[types.String], 2)}, true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[[2]string]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[[2]string]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[[2]string]",
		"testdata.TypedArray[[2]string]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[map[int]string]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewMap(types.Typ[types.Int], types.Typ[types.String])},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[map[int]string]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[map[int]string]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[map[int]string]",
		"testdata.TypedArray[map[int]string]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[chan error]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewChan(types.SendRecv, g.TType(reflect.TypeFor[error]()))},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[chan error]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[chan error]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[chan error]",
		"testdata.TypedArray[chan error]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[chan<- testdata.Tagged]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewChan(types.SendOnly, tTestdataTagged)},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[chan<- github.com/xoctopus/typex/testdata.Tagged]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[chan<- xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.Tagged]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[chan<- github.com/xoctopus/typex/testdata.Tagged]",
		"testdata.TypedArray[chan<- testdata.Tagged]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[<-chan *testdata.Tagged]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewChan(types.RecvOnly, types.NewPointer(tTestdataTagged))},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[<-chan *github.com/xoctopus/typex/testdata.Tagged]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[<-chan *xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.Tagged]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[<-chan *github.com/xoctopus/typex/testdata.Tagged]",
		"testdata.TypedArray[<-chan *testdata.Tagged]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[struct{}]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{tEmptyStruct},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[struct {}]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[struct {}]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[struct {}]",
		"testdata.TypedArray[struct {}]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[struct {
			A            string
			B            int
			testdata.Map `json:"esc''{}[]\""`
			testdata.TypedSlice[net.Addr]
			C struct {
				testdata.TypedArray[struct{ testdata.TypedSlice[net.Addr] }]
			}
		}]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewStruct(
				[]*types.Var{
					types.NewField(0, nil, "A", types.Typ[types.String], false),
					types.NewField(0, nil, "B", types.Typ[types.Int], false),
					types.NewField(0, p.Unwrap(), "Map", tTestdataMap, true),
					types.NewField(
						0, p.Unwrap(), "TypedSlice",
						resultx.Unwrap(types.Instantiate(nil, tTestdataTypedSlice, []types.Type{tNetAddr}, true)),
						true,
					),
					types.NewField(
						0, nil, "C",
						types.NewStruct([]*types.Var{
							types.NewField(
								0, p.Unwrap(), "TypedArray",
								resultx.Unwrap(types.Instantiate(nil, tTestdataTypedArray, []types.Type{
									types.NewStruct([]*types.Var{
										types.NewField(
											0, p.Unwrap(), "TypedSlice",
											resultx.Unwrap(types.Instantiate(
												nil, tTestdataTypedSlice,
												[]types.Type{tNetAddr}, true,
											)),
											true,
										),
									}, nil),
								}, true)),
								true,
							),
						}, nil),
						false,
					),
				},
				[]string{"", "", `json:"esc''{}[]\""`, ""},
			)},
			true,
		)),
		`github.com/xoctopus/typex/testdata.TypedArray[struct { ` +
			`A string; ` +
			`B int; ` +
			`github.com/xoctopus/typex/testdata.Map "json:\"esc''{}[]\\\"\""; ` +
			`github.com/xoctopus/typex/testdata.TypedSlice[net.Addr]; ` +
			`C struct { ` +
			`github.com/xoctopus/typex/testdata.TypedArray[struct { github.com/xoctopus/typex/testdata.TypedSlice[net.Addr] }] ` +
			`} ` +
			`}]`,
		`xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[struct { ` +
			`A string; ` +
			`B int; ` +
			`xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.Map "json:\"esc''{}[]\\\"\""; ` +
			`xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedSlice[net.Addr]; ` +
			`C struct { ` +
			`xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[struct { xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedSlice[net.Addr] }] ` +
			`} ` +
			`}]`,
		"github.com/xoctopus/typex/testdata",
		`TypedArray[struct { ` +
			`A string; ` +
			`B int; ` +
			`github.com/xoctopus/typex/testdata.Map "json:\"esc''{}[]\\\"\""; ` +
			`github.com/xoctopus/typex/testdata.TypedSlice[net.Addr]; ` +
			`C struct { ` +
			`github.com/xoctopus/typex/testdata.TypedArray[struct { github.com/xoctopus/typex/testdata.TypedSlice[net.Addr] }] ` +
			`} ` +
			`}]`,
		`testdata.TypedArray[struct { ` +
			`A string; ` +
			`B int; ` +
			`testdata.Map "json:\"esc''{}[]\\\"\""; ` +
			`testdata.TypedSlice[net.Addr]; ` +
			`C struct { ` +
			`testdata.TypedArray[struct { testdata.TypedSlice[net.Addr] }] ` +
			`} ` +
			`}]`,
	},
	{
		reflect.TypeFor[testdata.TypedArray[interface{}]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewInterfaceType(nil, nil)},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[interface {}]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[interface {}]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[interface {}]",
		"testdata.TypedArray[interface {}]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[interface {
			fmt.Stringer
			io.Closer
			io.WriteCloser
			io.ReadCloser
			io.ReadWriter
		}]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewInterfaceType(nil, []types.Type{tFmtStringer, tIoReadCloser, tIoWriteCloser, tIoReadWriter})},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }]",
		"testdata.TypedArray[interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[func()]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewSignatureType(nil, nil, nil, nil, nil, false)},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[func()]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[func()]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[func()]",
		"testdata.TypedArray[func()]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[func(any, ...string) string]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewSignatureType(
				nil, nil, nil,
				types.NewTuple(
					types.NewParam(0, nil, "", types.NewInterfaceType(nil, nil)),
					types.NewParam(0, nil, "", types.NewSlice(types.Typ[types.String])),
				),
				types.NewTuple(
					types.NewParam(0, nil, "", types.Typ[types.String]),
				),
				true,
			)},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[func(interface {}, ...string) string]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[func(interface {}, ...string) string]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[func(interface {}, ...string) string]",
		"testdata.TypedArray[func(interface {}, ...string) string]",
	},
	{
		reflect.TypeFor[testdata.TypedArray[func(any, ...string) (string, error)]](),
		resultx.Unwrap(types.Instantiate(
			nil, tTestdataTypedArray,
			[]types.Type{types.NewSignatureType(
				nil, nil, nil,
				types.NewTuple(
					types.NewParam(0, nil, "", types.NewInterfaceType(nil, nil)),
					types.NewParam(0, nil, "", types.NewSlice(types.Typ[types.String])),
				),
				types.NewTuple(
					types.NewParam(0, nil, "", types.Typ[types.String]),
					types.NewParam(0, nil, "", g.TType(reflect.TypeFor[error]())),
				),
				true,
			)},
			true,
		)),
		"github.com/xoctopus/typex/testdata.TypedArray[func(interface {}, ...string) (string, error)]",
		"xwrap_github_d_com_s_xoctopus_s_typex_s_testdata.TypedArray[func(interface {}, ...string) (string, error)]",
		"github.com/xoctopus/typex/testdata",
		"TypedArray[func(interface {}, ...string) (string, error)]",
		"testdata.TypedArray[func(interface {}, ...string) (string, error)]",
	},
}

func TestGlobal(t *testing.T) {
	t.Run("Wrap", func(t *testing.T) {
		for _, c := range GlobalCases {
			Expect(t, g.Wrap(c.rt), Equal(c.wrapped))
			Expect(t, g.Wrap(c.tt), Equal(c.wrapped))
		}
		t.Run("InvalidInput", func(t *testing.T) {
			ExpectPanic(
				t,
				func() { _ = g.Wrap("") },
				ErrorContains("invalid wrap key type"),
			)
		})
	})
	t.Run("Literalize", func(t *testing.T) {
		for _, c := range GlobalCases {
			if c.rt == nil {
				Expect(t, c.tt, BeNil[types.Type]())
				Expect(t, g.Literalize(c.rt), BeNil[internal.Literal]())
				Expect(t, g.Literalize(c.rt), BeNil[internal.Literal]())
				continue
			}
			ur := g.Literalize(c.rt)
			ut := g.Literalize(c.tt)

			Expect(t, reflect.DeepEqual(ur, ut), BeTrue())
			Expect(t, ur.String(), Equal(c.id))
			Expect(t, ur.PkgPath(), Equal(c.pkg))
			Expect(t, ur.Name(), Equal(c.name))
			Expect(t, ur.TypeLit(context.Background()), Equal(c.literal))

			if builtin, ok := ur.(internal.Builtin); ok {
				Expect(t, builtin.Kind(), Equal(c.rt.Kind()))
			}
		}
		t.Run("InvalidInput", func(t *testing.T) {
			ExpectPanic(
				t,
				func() { _ = g.Literalize("") },
				ErrorContains("invalid literalize key type"),
			)
		})
	})
	t.Run("TType", func(t *testing.T) {
		for _, c := range GlobalCases {
			if c.rt == nil {
				continue
			}
			rtt := g.TType(c.rt)
			utt := g.TType(g.Literalize(c.rt))
			identical := types.Identical(utt, rtt)
			Expect(t, identical, BeTrue())
		}
		t.Run("InvalidInput", func(t *testing.T) {
			ExpectPanic(
				t,
				func() { _ = g.TType("") },
				ErrorContains("invalid ttype key type"),
			)
		})
	})

	t.Run("Namer", func(t *testing.T) {
		ctx := namer.WithContext(context.Background(), &TestPackageNamer{})
		u := g.Literalize(reflect.TypeFor[testdata.TypedArray[testdata.Map]]())
		Expect(t, u.TypeLit(ctx), Equal("demo.TypedArray[demo.Map]"))
	})
}

type TestPackageNamer struct{}

func (t TestPackageNamer) Package(path string) string {
	if path == "github.com/xoctopus/typex/testdata" {
		return "demo"
	}
	return path
}
