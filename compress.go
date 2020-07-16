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

	if err = os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
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

	_, err = io.Copy(tw, f)
	return
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
