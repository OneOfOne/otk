package otk

import (
	"unicode"
)

type KeyVal struct {
	Key string
	Val string
}

func KeywordsToSearch(search string) (out []KeyVal) {
	FindKeywords(search, ':', func(kw, val string) {
		out = append(out, KeyVal{Key: kw, Val: val})
	})

	return out[:len(out):len(out)]
}

func FindKeywords(s string, sepChar rune, fn func(kw, val string)) {
	wordStart, sepStart := -1, -1
	inQuote := false

	process := func(i int) {
		if wordStart == -1 {
			return
		}
		if sepStart == -1 {
			fn("", s[wordStart:i])
		} else if sepStart == len(s)-1 {
			fn("", s[wordStart:])
		} else {
			kw := s[wordStart:sepStart]
			if s[sepStart+1] == '"' {
				sepStart++
			}
			fn(kw, s[sepStart+1:i])
		}

		wordStart, sepStart, inQuote = -1, -1, false
	}

	for i, c := range s {
		switch {
		case inQuote && c != '"':
			continue

		case c == '"':
			if inQuote {
				process(i)
				continue
			}

			if inQuote = true; wordStart == -1 {
				wordStart = i + 1
			} else if sepStart == -1 {
				process(i)
			}

		case c == sepChar:
			if sepStart == -1 && wordStart != -1 {
				sepStart = i
			}

		case unicode.IsSpace(c):
			if wordStart == -1 {
				wordStart = i + 1
			} else {
				process(i)
			}

		default:
			if wordStart == -1 {
				wordStart = i
			}
		}
	}

	process(len(s))
}
