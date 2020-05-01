package otk

import (
	"archive/tar"
	"bufio"
	"io"
	"os"
	"path/filepath"

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

type TarOptions struct {
	BufSize       int                                    // buffer size, default size is 2 MiB
	CompressFn    func(w io.Writer) io.WriteCloser       // returns a compressor, ex gzip.NewWriter
	FilterFn      func(path string, fi os.FileInfo) bool // return false to ignore files
	DeleteOnError bool
}

func TarFolder(folder, fp string, topts *TarOptions) (err error) {
	const defBufSize = 2 * 1024 * 1024

	if topts == nil {
		topts = &TarOptions{}
	}

	if err = os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
		return
	}

	var f *os.File
	if f, err = os.Create(fp + ".tmp"); err != nil {
		return
	}

	bsz := topts.BufSize
	if bsz < 1 {
		bsz = defBufSize
	}
	bw := bufio.NewWriterSize(f, bsz)

	defer func() {
		if err = MergeErrors(", ", err, bw.Flush(), f.Close()); err != nil && topts.DeleteOnError {
			err = MergeErrors(", ", err, os.Remove(fp))
		}
		if err == nil {
			err = os.Rename(fp+".tmp", fp)
		}
	}()

	var wc io.WriteCloser
	if topts.CompressFn != nil {
		wc = topts.CompressFn(bw)
		defer func() { err = MergeErrors(", ", err, wc.Close()) }()
	} else {
		wc = NopWriteCloser(bw)
	}

	tw := tar.NewWriter(wc)
	defer func() { err = MergeErrors(", ", err, tw.Close()) }()

	ffn := topts.FilterFn
	if ffn == nil {
		ffn = func(_ string, fi os.FileInfo) bool { return fi.IsDir() || fi.Mode().IsRegular() }
	}

	err = filepath.Walk(folder, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		p, _ := filepath.Rel(folder, path)
		if !ffn(p, fi) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		return AppendToTar(tw, path, p, fi)
	})

	return
}

// AppendToTar is a helper function for add a physical file to tar
func AppendToTar(tw *tar.Writer, fullPath, tarPath string, fi os.FileInfo) error {
	hdr, err := tar.FileInfoHeader(fi, tarPath)
	if err != nil {
		return err
	}
	hdr.Name = tarPath

	if err = tw.WriteHeader(hdr); err != nil {
		return xerrors.Errorf("writing header for %s: %w", tarPath, err)
	}

	f, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	f.Close()

	return err
}
