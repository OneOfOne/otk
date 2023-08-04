package otk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTarFolder(t *testing.T) {
	dp := t.TempDir()
	//	defer os.RemoveAll(dp)

	fp := filepath.Join(dp, "otk.tar")
	err := TarFolder(prefixPath /* full path to this package */, fp, &TarOptions{
		FilterFn: func(path string, fi os.FileInfo) bool {
			return fi.IsDir() || strings.HasSuffix(path, ".go")
		},
		DeleteOnError: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := Untar(fp, dp); err != nil {
		t.Fatal(err)
	}
	t.Log(filepath.Glob(filepath.Join(dp, "*")))
}
