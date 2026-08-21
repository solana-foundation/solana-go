package encryption_test

import (
	"testing"

	zk "github.com/gagliardetto/solana-go/programs/zk-elgamal-proof"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
)

func TestAeEncryptDecrypt(t *testing.T) {
	amount := zktest.GenAmount(t, 1<<63)
	key, err := encryption.NewAeKey()
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

	var wrongKey encryption.AeKey
	_, err = wrongKey.Decrypt(ct)
	zktest.ExpectStatusError(t, err, zk.DECRYPTION_ERROR)
}
