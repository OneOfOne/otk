package otk

import (
	"reflect"
	"testing"
)

func TestAES(t *testing.T) {
	data := []string{"uid", "ts", "something"}
	enc, err := AESEncrypt(data, "::", "pass", 0)
	if err != nil {
		t.Fatal(err)
	}

	parts, err := AESDecrypt(enc, "::", "pass", 0)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, parts) {
		t.Fatalf("expected %q, got %q", data, parts)
	}
}
