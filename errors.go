package otk

import (
	"strings"

	"golang.org/x/xerrors"
)

// MergeErrors merges a slice of errors, but they will lose any context.
// returns nil if all errors are nil
func MergeErrors(sep string, errs ...error) error {
	var buf strings.Builder

	for _, err := range errs {
		if err == nil {
			continue
		}

		if buf.Len() > 0 {
			buf.WriteString(sep)
		}

		buf.WriteString(err.Error())
	}

	if buf.Len() == 0 {
		return nil
	}

	return xerrors.New(buf.String())
}
