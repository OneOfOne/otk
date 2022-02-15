package otk

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/xerrors"
)

// PipeRd is an io.Pipe helper, returns a io.ReadCloser
func PipeRd(writeFn func(io.Writer) error) io.ReadCloser {
	rd, wr := io.Pipe()
	go func() { wr.CloseWithError(writeFn(wr)) }()
	return rd
}

// PipeWr is an io.Pipe helper, returns a io.WriteCloser
func PipeWr(readerFn func(io.Reader) error) io.WriteCloser {
	rd, wr := io.Pipe()
	go func() { rd.CloseWithError(readerFn(rd)) }()
	return wr
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
	var (
		p  []byte // lazily initialized if/when needed
		ln = len(s)
	)
	for i, w := range t.writers {
		if w == nil {
			continue
		}
		if sw, ok := w.(io.StringWriter); ok {
			n, err = sw.WriteString(s)
		} else {
			if p == nil {
				p = []byte(s)
				ln = len(p)
			}
			n, err = w.Write(p)
		}
		if err == nil && n != ln {
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

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func NopWriteCloser(w io.Writer) io.WriteCloser { return nopCloser{w} }

func CopyOnWriteFile(fp string, fn func(w io.Writer) error) (err error) {
	return CopyOnWriteFilePerms(fp, func(w *bufio.Writer) error { return fn(w) }, 0644)
}

var bufWriterPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewWriterSize(nil, 32768)
	},
}

func CopyOnWriteFilePerms(fp string, fn func(bw *bufio.Writer) error, mode os.FileMode) (err error) {
	var f *os.File
	dir, fname := filepath.Split(fp)
	if f, err = os.CreateTemp(dir, fname); err != nil {
		return
	}

	bw := bufWriterPool.Get().(*bufio.Writer)
	bw.Reset(f)

	defer func() {
		bw.Reset(nil)
		bufWriterPool.Put(bw)
		os.Remove(f.Name()) // clean our trash if we errored out
	}()

	if err = fn(bw); err != nil {
		f.Close()
		return
	}

	if err = bw.Flush(); err != nil {
		f.Close()
		return
	}

	if err = f.Chmod(mode); err != nil {
		f.Close()
		return
	}

	if err = f.Close(); err != nil {
		return
	}

	return os.Rename(f.Name(), fp)
}

type (
	FileDecoder interface {
		Decode(interface{}) error
	}
	FileEncoder interface {
		Encode(interface{}) error
	}
)

func ReadFileWithDecoder(fp string, dec func(r io.Reader) FileDecoder, out interface{}) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	return dec(f).Decode(out)
}

func WriteFileWithEncoder(fp string, enc func(w io.Writer) FileEncoder, in interface{}) error {
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	return enc(f).Encode(in)
}

func ReadJSONFile(fp string, out interface{}) error {
	return ReadFileWithDecoder(fp, func(r io.Reader) FileDecoder {
		return json.NewDecoder(r)
	}, out)
}

func WriteJSONFile(fp string, in interface{}, indent bool) error {
	return WriteFileWithEncoder(fp, func(w io.Writer) FileEncoder {
		enc := json.NewEncoder(w)
		if indent {
			enc.SetIndent("", "\t")
		}
		return enc
	}, in)
}

type CachedReader struct {
	R      io.Reader
	tmp    []byte
	rewind bool
}

func (cr *CachedReader) Read(p []byte) (int, error) {
	if !cr.rewind {
		if cr.tmp == nil {
			cr.tmp = make([]byte, 0, len(p))
		}
		n, err := cr.R.Read(p)
		cr.tmp = append(cr.tmp, p[:n]...)
		return n, err
	}

	if len(cr.tmp) > 0 {
		n := copy(p, cr.tmp)
		if cr.tmp = cr.tmp[n:]; len(cr.tmp) == 0 {
			cr.tmp = nil
		}
		return n, nil
	}

	return cr.R.Read(p)
}

func (cr *CachedReader) Rewind() {
	cr.rewind = true
}
