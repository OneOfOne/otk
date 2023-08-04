package otk

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/xerrors"
)

type TarOptions struct {
	CompressFn    func(w io.Writer) io.WriteCloser
	UncompressFn  func(r io.Reader) io.Reader
	FilterFn      func(path string, fi os.FileInfo) bool
	BufSize       int
	DeleteOnError bool
}

func TarFolder(folder, fp string, opts *TarOptions) (err error) {
	if opts == nil {
		opts = &TarOptions{}
	}

	if err = os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
		return
	}

	var f *os.File
	if f, err = os.Create(fp + ".tmp"); err != nil {
		return
	}

	defer func() {
		if err = MergeErrors(", ", err, f.Close()); err != nil && opts.DeleteOnError {
			err = MergeErrors(", ", err, os.Remove(fp))
		}
		if err == nil {
			err = os.Rename(fp+".tmp", fp)
		}
	}()
	return Tar(folder, f, opts)
}

func Tar(folder string, w io.Writer, opts *TarOptions) (err error) {
	const defBufSize = 4 * 1024 * 1024

	if opts == nil {
		opts = &TarOptions{}
	}

	bsz := opts.BufSize
	if bsz < 1 {
		bsz = defBufSize
	}
	bw := bufio.NewWriterSize(w, bsz)
	defer func() { err = MergeErrors(", ", err, bw.Flush()) }()

	var wc io.WriteCloser
	if opts.CompressFn != nil {
		wc = opts.CompressFn(bw)
		defer func() { err = MergeErrors(", ", err, wc.Close()) }()
	} else {
		wc = NopWriteCloser(bw)
	}

	tw := tar.NewWriter(wc)
	defer func() { err = MergeErrors(", ", err, tw.Close()) }()

	ffn := opts.FilterFn
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

		if err = AppendToTar(tw, path, p); err != nil {
			err = xerrors.Errorf("tar error (%s): %w", path, err)
		}
		return err
	})

	return
}

// AppendToTar is a helper function for add a physical file to tar
func AppendToTar(tw *tar.Writer, fullPath, tarPath string) (err error) {
	var (
		f   *os.File
		st  os.FileInfo
		hdr *tar.Header
	)
	if f, err = os.Open(fullPath); err != nil {
		return err
	}
	defer f.Close()

	if st, err = f.Stat(); err != nil {
		return
	}

	if hdr, err = tar.FileInfoHeader(st, tarPath); err != nil {
		return
	}
	hdr.Name = tarPath

	if err = tw.WriteHeader(hdr); err != nil {
		return
	}

	_, err = io.Copy(tw, io.LimitReader(f, st.Size()))
	return
}

func UntarFolder(fp, folder string, opts *TarOptions) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()
	return Untar(f, folder, opts)
}

func Untar(r io.Reader, folder string, opts *TarOptions) error {
	r = bufio.NewReader(r)
	if opts != nil && opts.UncompressFn != nil {
		r = opts.UncompressFn(r)
	}
	rd := tar.NewReader(r)

	for {
		hdr, err := rd.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		p := filepath.Join(folder, hdr.Name)

		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}

		if err = CopyOnWriteFile(p, func(w io.Writer) error {
			_, err := io.Copy(w, rd)
			return err
		}); err != nil {
			return err
		}

	}
}

func Unzip(rt io.ReaderAt, dst string, filter func(path string, f *zip.File) bool) (err error) {
	var (
		zr   *zip.Reader
		size int64
	)

	switch rt := rt.(type) {
	case interface{ Len() int }:
		size = int64(rt.Len())
	case interface{ Size() int64 }:
		size = rt.Size()
	case interface{ Stat() (os.FileInfo, error) }:
		fi, err := rt.Stat()
		if err != nil {
			return err
		}
		size = fi.Size()
	default:
		err = xerrors.Errorf("%T doesn't provide a way to get the file size", rt)
		return
	}

	if zr, err = zip.NewReader(rt, size); err != nil {
		return
	}

	for _, zf := range zr.File {
		fpath := filepath.Join(dst, zf.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dst)+string(os.PathSeparator)) {
			err = xerrors.Errorf("%s: illegal file path", fpath)
			return
		}

		if filter != nil && !filter(fpath, zf) {
			continue
		}

		if zf.FileInfo().IsDir() {
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return
		}

		func() {
			var f *os.File
			if f, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode()); err != nil {
				return
			}
			defer f.Close()

			var rc io.ReadCloser
			if rc, err = zf.Open(); err != nil {
				return
			}
			defer rc.Close()

			_, err = io.Copy(f, rc)
		}()

		if err != nil {
			return xerrors.Errorf("%s: unzip error: %w", zf.Name, err)
		}
	}

	return
}
