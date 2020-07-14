package otk

import (
	"regexp"
	"sync"
)

var (
	emails     *regexp.Regexp
	emailsOnce sync.Once
)

// Email returns a copy
func Email() *regexp.Regexp { return emails.Copy() }

// MatchEmail matches against the main regexp, not recommended for high concurrency.
func MatchEmail(email string) bool {
	emailsOnce.Do(func() {
		// source http://emailregex.com/ javascript regex
		emails = regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)
	})

	return emails.MatchString(email)
}
