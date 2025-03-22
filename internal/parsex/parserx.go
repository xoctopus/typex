package parsex

import (
	"slices"
	"strconv"
	"strings"

	"github.com/xoctopus/x/misc/must"
)

var brackets = map[rune]rune{
	'(': ')',
	'[': ']',
	'{': '}',
}

// Bracketed returns the sub string of `id` bracketed by `identifier0` and the indexes
// of brackets
func Bracketed(id string, identifier0 rune) (string, int, int) {
	identifier1, ok := brackets[identifier0]
	must.BeTrueWrap(ok, "invalid bracket identifier: %v", identifier0)

	l, r, embeds, quoted := -1, -1, 0, false
End:
	for i := 0; i < len(id); i++ {
		c := rune(id[i])
		switch c {
		case identifier0:
			if quoted {
				continue
			}
			if embeds == 0 {
				l = i
			}
			embeds++
		case identifier1:
			if quoted {
				continue
			}
			embeds--
			if embeds == 0 {
				must.BeTrue(l >= 0)
				r = i
				break End
			}
		case '"':
			quoted = !quoted
		case '\\':
			i++
		}
	}
	must.BeTrue(l >= 0 && r > 0 || l == -1 && r == -1)
	if l < 0 && r < 0 {
		return "", -1, -1
	}
	return id[l+1 : r], l, r
}

// quoted returns the quoted sub string and quoter indexes.
func quoted(id string) (string, int, int) {
	quoted, l, r := false, -1, -1
	for i, c := range id {
		if i < len(id)-1 && id[i+1] == '\\' {
			i += 2
			continue
		}
		if c == '"' {
			if !quoted && l < 0 {
				quoted = true
				l = i
				continue
			}
			if quoted && l >= 0 {
				r = i
				break
			}
		}
	}
	return id[l+1 : r], l, r
}

// Separate separates id by `sep`, unlike strings.Split, it will ignore `sep` within
// code blocks. eg:
// id = `struct { A string }`; sep = ' ', returns `struct`, `{ A string }`, sep
// in struct block will be ignored
// id = `path/to/pkg.Typename[T1,T2]`; sep = ',', returns `path/to/pkg.Typename[T1,T2]`,
// sep `,` in type argument list will be ignored
// id = ` A string; B int; C string "json:\"c,omitempty\"" `; sep = ';', returns
// `A string`, `B int` and `C string "json:\"c,omitempty\"`
func Separate(id string, sep rune) []string {
	must.BeTrue(sep == ',' || sep == ';' || sep == ' ')

	if len(id) == 0 {
		return nil
	}

	var (
		parts   = make([]string, 0)
		part    []rune
		embeds  = map[rune]int{'(': 0, '{': 0, '[': 0}
		quoting bool
	)

	for i := 0; i < len(id); i++ {
		c := rune(id[i])
		switch c {
		case '(':
			embeds['(']++
		case ')':
			embeds['(']--
			must.BeTrue(embeds['('] >= 0)
		case '[':
			embeds['[']++
		case ']':
			embeds['[']--
			must.BeTrue(embeds['['] >= 0)
		case '{':
			embeds['{']++
		case '}':
			embeds['{']--
			must.BeTrue(embeds['{'] >= 0)
		case '"':
			quoting = !quoting
		case sep:
			if embeds['('] == 0 && embeds['{'] == 0 && embeds['['] == 0 && !quoting {
				goto FinishPart
			}
		}
		part = append(part, c)
		if c == '\\' {
			part = append(part, rune(id[i+1]))
			i++
		}
		if i == len(id)-1 {
			goto FinishPart
		}
		continue
	FinishPart:
		p := strings.TrimSpace(string(part))
		must.BeTrue(len(p) > 0)
		parts = append(parts, p)
		part = part[0:0]
	}

	return parts
}

// reverse return reversed string
func reverse(id string) string {
	_id := []rune(id)
	slices.Reverse(_id)
	return string(_id)
}

// FieldInfo parses the struct field info, returns name, type and tag of field
// input `id` gives from reflect.Type.String()
// When a struct type appears in the parameter list of a generic type, there are
// the following special cases to handle input `id`
// 1. for unexported fields, the package path of the unexported field will be
// automatically included.
// eg: T[struct { a int }], the `id` will be
// PackageName.T[struct { FullPackagePathOfA.a int }]
// 2. for anonymous generic fields, it will be expressed as: TypeName = FullPackagePath.TypeName[TypeArgs]
// eg: T[struct { T2[int] }]
// PackageNameOfT.T[struct { T2 = FullPackagePathOfT2.T2[int] }]
// see Example_structInTypeArguments
func FieldInfo(id string) (name string, typ string, tag string) {
	if id[len(id)-1] == '"' {
		id = reverse(id)

		_tag, ql, qr := quoted(id)
		must.BeTrue(ql >= 0 && qr > 0)
		_tag = "\"" + reverse(_tag) + "\""
		_tag, err := strconv.Unquote(_tag)
		must.NoError(err)
		tag = _tag

		id = strings.TrimSpace(reverse(id[qr+1:]))
	}

	switch parts := Separate(id, ' '); len(parts) {
	case 1:
		typ = parts[0]
	case 2:
		name, typ = parts[0], parts[1]
	default:
		if len(parts) == 3 && parts[1] == "=" {
			typ = parts[2]
		} else {
			name, typ = parts[0], strings.Join(parts[1:], " ")
		}
	}
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return
}
