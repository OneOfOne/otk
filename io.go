package otk

import (
	"io"

	"golang.org/x/xerrors"
)

// Pipe is an io.Pipe helper, returns a io.ReadCloser
func Pipe(writeFn func(io.Writer) error) io.ReadCloser {
	rd, rw := io.Pipe()
	go func() { rw.CloseWithError(writeFn(rw)) }()
	return rd
}

type multiWriter struct {
	writers []io.Writer
	errs    []error
}

func (t *multiWriter) Write(p []byte) (n int, err error) {
	for i, w := range t.writers {
		if w == nil {
			continue
		}
		n, err = w.Write(p)
		if err == nil && n != len(p) {
			err = io.ErrShortWrite
		}
		if err != nil {
			t.errs = append(t.errs, xerrors.Errorf("writer #%d: %w", i, err))
			t.writers[i] = nil
		}
	}
	return len(p), nil
}

var _ io.StringWriter = (*multiWriter)(nil)

func (t *multiWriter) WriteString(s string) (n int, err error) {
	var p []byte // lazily initialized if/when needed
	for i, w := range t.writers {
		if w == nil {
			continue
		}
		if sw, ok := w.(io.StringWriter); ok {
			n, err = sw.WriteString(s)
		} else {
			if p == nil {
				p = []byte(s)
			}
			n, err = w.Write(p)
		}
		if err == nil && n != len(p) {
			err = io.ErrShortWrite
		}
		if err != nil {
			t.errs = append(t.errs, xerrors.Errorf("writer #%d: %w", i, err))
			t.writers[i] = nil
		}
	}
	return len(s), nil
}

func (t *multiWriter) getErrs() []error { return t.errs }

// MultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
//
// Each write is written to each listed writer, one at a time.
// If a listed writer returns an error, that writer will be ignored
// for later writes.
// returns the multi writer and a func that returns a slice of errors if there were any.
func MultiWriter(writers ...io.Writer) (io.Writer, func() []error) {
	allWriters := make([]io.Writer, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*multiWriter); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}
	mw := &multiWriter{allWriters, nil}
	return mw, mw.getErrs
}
