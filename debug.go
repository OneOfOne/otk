package otk

import (
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

func init() {
	initRelative()
}

var prefixPath string

func initRelative() {
	_, fileName, _, _ := runtime.Caller(0)
	log.Println(fileName)
	prefixPath = filepath.Dir(fileName)
}

// Caller returns function, file and line of the caller.
func Caller(skip int, trim bool) (function, file string, line int) {
	var uframes [3]uintptr
	runtime.Callers(skip+1, uframes[:])
	frames := runtime.CallersFrames(uframes[:])
	if _, ok := frames.Next(); !ok {
		return "", "", 0
	}

	fr, ok := frames.Next()
	if !ok {
		return
	}

	function, file, line = fr.Function, fr.File, fr.Line
	if trim {
		if idx := strings.LastIndexByte(function, '/'); idx != -1 {
			function = function[idx+1:]
		}
		file = strings.TrimPrefix(file, prefixPath)
	}
	return
}
