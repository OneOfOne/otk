package otk

import (
	"reflect"
	"testing"
	"time"
)

func TestAES(t *testing.T) {
	data := []string{"uid", "ts", "something", time.Now().UTC().String()}
	enc, err := AESEncrypt(data, "::", "pass", 16)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("encrypted data (%d): %s", len(enc), enc)

	parts, err := AESDecrypt(enc, "::", "pass", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("decrypted data: %q", parts)

	if !reflect.DeepEqual(data, parts) {
		t.Fatalf("expected %q, got %q", data, parts)
	}
}
