package otk

import (
	"strings"
	"sync"

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

type ErrorList []error

func (el ErrorList) Error() string { return MergeErrors(" | ", el...).Error() }
func (el ErrorList) Len() int      { return len(el) }

func (el ErrorList) Err() error {
	if len(el) == 0 {
		return nil
	}
	return el[:len(el):len(el)]
}

func (el *ErrorList) Push(errs ...error) {
	for _, err := range errs {
		if err != nil {
			*el = append(*el, err)
		}
	}
}

type SafeErrorList struct {
	mux sync.Mutex
	el  ErrorList
}

func (el *SafeErrorList) Error() string {
	el.mux.Lock()
	err := el.el.Error()
	el.mux.Unlock()
	return err
}

func (el *SafeErrorList) Len() int {
	el.mux.Lock()
	ln := len(el.el)
	el.mux.Unlock()
	return ln
}

func (el *SafeErrorList) Err() error {
	el.mux.Lock()
	err := el.el.Err()
	el.mux.Unlock()
	return err
}

func (el *SafeErrorList) Push(errs ...error) {
	el.mux.Lock()
	el.el.Push(errs...)
	el.mux.Unlock()
}
