package otk

import (
	"regexp"
	"strings"
	"sync"
)

var (
	emails     *regexp.Regexp
	emailsOnce sync.Once
)

// ValidEmail checks the given email for validity
func ValidEmail(email string) bool {
	emailsOnce.Do(func() {
		// source http://emailregex.com/ javascript regex
		emails = regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)
	})

	return emails.MatchString(email)
}

// ReplaceAllStringSubmatchFunc is a helper function to replace regexp sub matches.
// based on https://gist.github.com/slimsag/14c66b88633bd52b7fa710349e4c6749 (MIT)
// Note: slice `in` is reused internally, make a copy if you need to keep it, ex:
// 	cp := append([]string(nil), in...)
// example:
//	re := regexp.MustCompile(`([:*].*?)(?:/|$)`)
//	ReplaceAllStringSubmatchFunc(re, "/path/:id/:name", func(in []string) []string {
//		for i, s := range in {
//			in[i] = "my-" + s[1:]
//		}
//		return in
//	}, -1) === "/path/my-id/my-name"
func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, src string, repl func([]string) []string, n int) string {
	var (
		matches = re.FindAllStringSubmatchIndex(src, n)

		res               strings.Builder
		groups            []string
		groupIndices      [][2]int
		gi                [2]int
		start, end, last  int
		lastGroup, gs, ge int
	)

	res.Grow(len(src))

	for _, match := range matches {
		start, end = match[0], match[1]
		res.WriteString(src[last:start])
		last = end

		// Determine the groups / submatch bytes and indices.
		groups, groupIndices = groups[:0], groupIndices[:0]
		for i := 2; i < len(match); i += 2 {
			start := match[i]
			end := match[i+1]
			groups = append(groups, src[start:end])
			groupIndices = append(groupIndices, [2]int{start, end})
		}

		groups = repl(groups)

		// Append match data.
		lastGroup = start
		for i := range groups {
			gi = groupIndices[i]
			gs, ge = gi[0], gi[1]
			res.WriteString(src[lastGroup:gs])
			lastGroup = ge

			// Append the new group value.
			res.WriteString(groups[i])
		}
		res.WriteString(src[lastGroup:end])
	}

	res.WriteString(src[last:])

	return res.String()
}
