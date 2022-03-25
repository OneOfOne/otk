package otk_test

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"go.oneofone.dev/otk/v2"
)

func TestTar(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller((0))
	basePath := filepath.Dir(thisFile)
	tmp, _ := os.CreateTemp("", "")
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

func TestCoW(t *testing.T) {
	tmp, _ := os.CreateTemp("", "")
	tmp.Close()
	defer os.Remove(tmp.Name())

	if err := otk.CopyOnWriteFilePerms(tmp.Name(), func(bw *bufio.Writer) error {
		_, err := bw.WriteString("test")
		return err
	}, 0o644); err != nil {
		t.Fatal(err)
	}

	if b, _ := os.ReadFile(tmp.Name()); string(b) != "test" {
		t.Fatalf("expected `test`, got %q", b)
	}

	fi, err := os.Stat(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if m := fi.Mode(); m != 0o644 {
		t.Fatalf("expected 0644, got 0%o", m)
	}
}
