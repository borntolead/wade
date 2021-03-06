package main

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/gowade/whtml"
)

// textPart represents either a typical HTML text node or a {{mustache node}}
type textPart struct {
	content    string
	isMustache bool
}

var (
	MustacheRegex = regexp.MustCompile("{{((?:[^{}]|{[^{]|}[^}])+)}}")
)

func escapeNewlines(str string) string {
	var buf bytes.Buffer
	for _, c := range []rune(str) {
		if c == '\n' {
			buf.WriteString(`\n`)
		} else {
			buf.WriteRune(c)
		}
	}

	return buf.String()
}

// parseTextMustache splits HTML text into a list of text and mustaches.
//
// "ABC: {{mustache}} DEF" would be splitted into
// "ABC: ", {{mustache}} and "DEF"
func parseTextMustache(text string) []textPart {
	matches := MustacheRegex.FindAllStringSubmatch(text, -1)

	if matches == nil {
		return []textPart{{text, false}}
	}

	parts := []textPart{}
	splitted := MustacheRegex.Split(text, -1)

	for i, m := range matches {
		if splitted[i] != "" {
			parts = append(parts, textPart{splitted[i], false})
		}

		parts = append(parts, textPart{strings.TrimSpace(m[1]), true})
	}

	if splitted[len(splitted)-1] != "" {
		parts = append(parts, textPart{splitted[len(splitted)-1], false})
	}

	return parts
}

// attributeValueCode returns the Go code that represents a string,
// formatted according to the mustaches in the value
func interpStrValueCode(fmtStr string, mustaches []string) string {
	mStr := strings.Join(mustaches, ", ")
	return sfmt(`fmt.Sprintf(%v, %v)`, fmtStr, mStr)
}

func valueToStrCode(value string) string {
	return sfmt(`wade.Str(%v)`, value)
}

// attributeValueCode returns the Go code that represents either a string or
// a single mustache value
func attributeValueCode(attr whtml.Attribute) string {
	switch attr.Type {
	case whtml.BoolAttribute:
		return "true"
	case whtml.StringAttribute:
		if len(attr.Mustaches) == 0 {
			return sfmt("`%v`", attr.Val)
		}
		return interpStrValueCode(attr.Val, attr.Mustaches)
	case whtml.MustacheAttribute:
		return attr.Val
	}

	panic("Unhandled attribute type")
	return ""
}

func justPeskySpaces(str string) bool {
	for _, c := range str {
		switch c {
		case '\n', '\t', ' ':
		default:
			return false
		}
	}

	return true
}
