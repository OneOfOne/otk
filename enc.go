package otk

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

var encBP BufferPool

func hashKey(key string, keySize uint8) []byte {
	switch keySize {
	case 16, 24, 32:
	default:
		keySize = 16
	}
	h := sha256.Sum256([]byte(key))
	return h[:keySize:keySize]
}

// AESEncrypt joins parts using sep, encrypts and seals it and returns b64 string.
// valid sizes are 16, 24 and 32 for AES-128, 192 and 256 respectively, defaults to 16.
func AESEncrypt(parts []string, sep string, passphrase string, keySize uint8) (_ string, err error) {
	var (
		block cipher.Block
		gcm   cipher.AEAD
		sz    = len(sep) * len(parts)
		buf   = encBP.Get()
	)

	defer encBP.Put(buf)

	if block, err = aes.NewCipher(hashKey(passphrase, keySize)); err != nil {
		return
	}
	if gcm, err = cipher.NewGCM(block); err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	for _, p := range parts {
		sz += len(p)
	}

	buf.Grow(sz)

	for i, p := range parts {
		if i > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(p)
	}

	enc := gcm.Seal(nonce, nonce, buf.Bytes(), nil)
	return base64.RawURLEncoding.EncodeToString(enc), nil
}

func AESDecrypt(b64Data, sep, passphrase string, keySize uint8) (parts []string, err error) {
	var (
		block cipher.Block
		gcm   cipher.AEAD
		data  []byte
		plain []byte
	)

	if data, err = base64.RawURLEncoding.DecodeString(b64Data); err != nil {
		return
	}

	if block, err = aes.NewCipher(hashKey(passphrase, keySize)); err != nil {
		return
	}

	if gcm, err = cipher.NewGCM(block); err != nil {
		return
	}

	nsz := gcm.NonceSize()

	if len(data) <= nsz {
		err = fmt.Errorf("invalid input data (decoded len: %d): %s", len(data), b64Data)
		return
	}

	if plain, err = gcm.Open(nil, data[:nsz], data[nsz:], nil); err != nil {
		return
	}

	parts = strings.Split(UnsafeString(plain), sep)
	return
}

func RandomString(sz int) (string, error) {
	b := make([]byte, 1+(sz/2))
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b)[:sz], nil
}
