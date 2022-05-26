package otk

import (
	"unicode"
)

func KeywordsToSearch(search string) (kws map[string]Set, rest Set) {
	MakeKeywords(search, ':', func(kw, val string) {
		if kw == "" {
			rest = rest.Add(val)
		} else {
			if kws == nil {
				kws = map[string]Set{}
			}
			kws[kw] = kws[kw].Add(val)
		}
	})

	return
}

func MakeKeywords(s string, sepChar rune, fn func(kw, val string)) {
	wordStart, sepStart := -1, -1
	inQuote := false

	process := func(i int) {
		if wordStart == -1 {
			return
		}
		if sepStart == -1 {
			fn("", s[wordStart:i])
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
