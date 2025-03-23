package parsex_test

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xoctopus/x/resultx"

	"github.com/xoctopus/typex/internal/parsex"
)

var cases = []struct {
	id      string
	bracket rune
	sep     rune
	parts   []string
}{
	{
		`path/to/pkg.TypeName[TypeArg1,TypeArg2[TArg2_1,TArg2_2]]`,
		'[', ',',
		[]string{"TypeArg1", "TypeArg2[TArg2_1,TArg2_2]"},
	},
	{
		`struct { A string "json:\"\\{}[]()\\\"\""; B int }`,
		'{', ';',
		[]string{
			`A string "json:\"\\{}[]()\\\"\""`,
			"B int",
		},
	},
	{
		`func(` +
			`int, ` +
			`struct { F func(); T TypeName[Ta1,Ta2] "json:\"t,omitempty\"" }, ` +
			`path/to/pkg.TypeName[Ta1,Ta2[struct { T T[Ta1,Ta2] "json:\"t\""; F func() }]], ` +
			`func() bool, ` +
			`...any` +
			`) (any, bool)`,
		'(', ',',
		[]string{
			"int",
			`struct { F func(); T TypeName[Ta1,Ta2] "json:\"t,omitempty\"" }`,
			`path/to/pkg.TypeName[Ta1,Ta2[struct { T T[Ta1,Ta2] "json:\"t\""; F func() }]]`,
			"func() bool",
			"...any",
		},
	},
	{
		"",
		'(', ',',
		nil,
	},
}

func TestBracketedAndSeparate(t *testing.T) {
	for _, c := range cases {
		sub, l, r := parsex.Bracketed(c.id, c.bracket)
		parts := parsex.Separate(sub, c.sep)
		NewWithT(t).Expect(parts).To(Equal(c.parts))
		if sub != "" {
			NewWithT(t).Expect(c.id[l+1 : r]).To(Equal(sub))
		} else {
			NewWithT(t).Expect(l).To(Equal(-1))
			NewWithT(t).Expect(r).To(Equal(-1))
		}
	}
}

type TT[T any] struct{}

func Example_structInTypeArguments() {
	t1 := reflect.TypeOf(TT[struct {
		string  // unexported field
		TT[int] // embedded generic type field
	}]{})
	fmt.Println(t1.String())

	targs0 := resultx.ResultsOf(parsex.Bracketed(t1.String(), '[')).At(0).(string)
	fields := resultx.ResultsOf(parsex.Bracketed(targs0, '{')).At(0).(string)

	for i, f := range parsex.Separate(fields, ';') {
		name, typ, tag := parsex.FieldInfo(f)
		fmt.Printf("field%d: name=%s;type=%s;tag=%s\n", i, name, typ, tag)
	}

	// Output:
	// parsex_test.TT[struct { github.com/xoctopus/typex/internal/parsex_test.string = string; TT = github.com/xoctopus/typex/internal/parsex_test.TT[int] }]
	// field0: name=;type=string;tag=
	// field1: name=;type=github.com/xoctopus/typex/internal/parsex_test.TT[int];tag=
}

func TestFieldInfo(t *testing.T) {
	t.Run("StructInTypeArg", func(t *testing.T) {
		tt := reflect.TypeOf(TT[struct {
			string
			A       int `json:"a,\"'{}()[]//\\"`
			TT[int] `json:"tt"`
			TT2     TT[struct{ A int }]
			Reader  io.Reader
			a       struct{ A string }
			AA      struct{ B int }
		}]{})

		fields := resultx.ResultsOf(parsex.Bracketed(
			resultx.ResultsOf(parsex.Bracketed(tt.String(), '[')).At(0).(string),
			'{',
		)).At(0).(string)

		expects := [][3]string{
			{"", "string", ""},
			{"A", "int", `json:"a,\"'{}()[]//\\"`},
			{"", "github.com/xoctopus/typex/internal/parsex_test.TT[int]", `json:"tt"`},
			{"TT2", "github.com/xoctopus/typex/internal/parsex_test.TT[struct { A int }]", ""},
			{"Reader", "io.Reader", ""},
			{"a", "struct { A string }", ""},
			{"AA", "struct { B int }", ""},
		}

		for i, f := range parsex.Separate(fields, ';') {
			name, typ, tag := parsex.FieldInfo(f)
			NewWithT(t).Expect(name).To(Equal(expects[i][0]))
			NewWithT(t).Expect(typ).To(Equal(expects[i][1]))
			NewWithT(t).Expect(tag).To(Equal(expects[i][2]))
		}
	})
	t.Run("Struct", func(t *testing.T) {
		tt := reflect.TypeOf(struct {
			string
			A       int `json:"a,\"'{}()[]//\\"`
			TT[int] `json:"tt"`
			TT2     TT[struct{ A int }]
			Reader  io.Reader
			a       struct{ A string }
			AA      struct{ B int }
		}{})

		fields := resultx.ResultsOf(parsex.Bracketed(tt.String(), '{')).At(0).(string)

		expects := [][3]string{
			{"", "string", ""},
			{"A", "int", `json:"a,\"'{}()[]//\\"`},
			{"", "parsex_test.TT[int]", `json:"tt"`},
			{"TT2", "parsex_test.TT[struct { A int }]", ""},
			{"Reader", "io.Reader", ""},
			{"a", "struct { A string }", ""},
			{"AA", "struct { B int }", ""},
		}

		for i, f := range parsex.Separate(fields, ';') {
			name, typ, tag := parsex.FieldInfo(f)
			NewWithT(t).Expect(name).To(Equal(expects[i][0]))
			NewWithT(t).Expect(typ).To(Equal(expects[i][1]))
			NewWithT(t).Expect(tag).To(Equal(expects[i][2]))
		}
	})
}
