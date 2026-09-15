package encryption_test

import (
	"testing"

	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
	zk "github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/zktest"
)

func TestAeEncryptDecrypt(t *testing.T) {
	amount := zktest.GenAmount(t, 1<<63)
	key, err := encryption.NewAeKey()
	if err != nil {
		t.Fatal(err)
	}
	ct, err := encryption.AeEncrypt(key, amount)
	if err != nil {
		t.Fatal(err)
	}
	got, err := encryption.AeDecrypt(key, ct)
	if err != nil {
		t.Fatal(err)
	}
	if got != amount {
		t.Fatalf("decrypted %d, want %d", got, amount)
	}

	var wrongKey zkencryption.AeKey
	_, err = encryption.AeDecrypt(wrongKey, ct)
	zktest.ExpectStatusError(t, err, zk.DECRYPTION_ERROR)
}
