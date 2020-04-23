package otk_test

import (
	"go/build"
	"log"
	"runtime"
	"testing"

	"github.com/OneOfOne/otk"
)

func TestCaller(t *testing.T) {
	log.Printf("%#+v", build.Default)
	t.Log(otk.Caller(0, true))
	t.Log(runtime.Caller(0))
}
