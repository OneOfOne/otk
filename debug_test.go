package otk_test

import (
	"go/build"
	"runtime"
	"testing"

	"go.oneofone.dev/otk/v2"
)

func TestCaller(t *testing.T) {
	t.Logf("%#+v", build.Default)
	t.Log(otk.Caller(0, true))
	t.Log(otk.Caller(0, false))
	t.Log(runtime.Caller(0))
}
