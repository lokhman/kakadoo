package app

import (
	"strings"
	"unicode/utf8"
)

func StripHtmlTags(s string) string {
	var builder strings.Builder

	builder.Grow(len(s) + utf8.UTFMax)

	in := false
	start := 0
	end := 0
	for i, c := range s {
		if i == len(s)-1 && end >= start {
			builder.WriteString(s[end:])
		}
		if c == '<' {
			if !in {
				start = i
			}
			in = true

			builder.WriteString(s[end:start])
			continue
		} else if c != '>' {
			continue
		}
		in = false
		end = i + 1
	}
	return builder.String()
}
