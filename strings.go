package otk

import (
	"regexp"
	"strings"
)

var searchArgsRe = regexp.MustCompile(`"([^"]+)"|[^\s]+`)

func SplitSearch(search string, splitChar byte) (kws map[string]Set, rest Set) {
	sp := searchArgsRe.FindAllStringSubmatch(search, -1)
	if len(sp) == 0 {
		rest = NewSet(search)
		return
	}

	for _, p := range sp {
		kw := p[1]
		if kw == "" {
			kw = p[0]
		}

		idx := strings.IndexByte(kw, splitChar)
		if idx == -1 {
			rest = rest.Add(kw)
			continue
		}

		if kws == nil {
			kws = map[string]Set{}
		}

		kws[kw[:idx]] = kws[kw[:idx]].Add(kw[idx+1:])
	}

	return
}
