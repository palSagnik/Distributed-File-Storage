package enc

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	input := "Bar not Foooo"
	src := bytes.NewReader([]byte(input))
	dst := new(bytes.Buffer)

	key := NewEncryptionKey()
	_, err := StreamEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	out := new(bytes.Buffer)
	if _, err := StreamDecrypt(key, dst, out); err != nil {
		t.Error(err)
	}

	if out.String() != input {
		t.Error("decrypted message does not match input")
	}
	fmt.Println(out.String())
}