package otk

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTarFolder(t *testing.T) {
	dp, err := ioutil.TempDir("", "compress")
	if err != nil {
		t.Fatal(err)
	}
	//	defer os.RemoveAll(dp)

	fp := filepath.Join(dp, "otk.tgz")
	err = TarFolder(prefixPath /* full path to this package */, fp, &TarOptions{
		CompressFn: func(w io.Writer) io.WriteCloser { return gzip.NewWriter(w) },
		FilterFn: func(path string, fi os.FileInfo) bool {
			return fi.IsDir() || strings.HasSuffix(path, ".go")
		},
		DeleteOnError: true,
	})

	if err != nil {
		t.Fatal(err)
	}
}
