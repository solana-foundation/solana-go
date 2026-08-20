package zk

import (
	"testing"
)

func TestAeEncryptDecrypt(t *testing.T) {
	amount := genAmount(t, 1<<63)
	key, err := NewAeKey()
	if err != nil {
		t.Fatal(err)
	}
	ct, err := key.Encrypt(amount)
	if err != nil {
		t.Fatal(err)
	}
	got, err := key.Decrypt(ct)
	if err != nil {
		t.Fatal(err)
	}
	if got != amount {
		t.Fatalf("decrypted %d, want %d", got, amount)
	}

	var wrongKey AeKey
	_, err = wrongKey.Decrypt(ct)
	expectStatusError(t, err, DECRYPTION_ERROR)
}
