package otk_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"go.oneofone.dev/otk"
)

func TestTar(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller((0))
	basePath := filepath.Dir(thisFile)
	tmp, _ := ioutil.TempFile("", "")
	tmp.Close()
	defer os.Remove(tmp.Name())

	_ = basePath
	err := otk.TarFolder(basePath, tmp.Name(), &otk.TarOptions{
		CompressFn: func(w io.Writer) io.WriteCloser { return gzip.NewWriter(w) },
		FilterFn: func(path string, _ os.FileInfo) bool {
			return path != ".git"
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command("tar", "-tvzf", tmp.Name()).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(out, []byte("io_test.go")) || bytes.Contains(out, []byte(".git")) {
		t.Fatalf("unexpected tar output: %s", out)
	}
}
