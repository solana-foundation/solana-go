package confidential

import (
	"testing"

	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/zktest"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

func generateSourceAccount(t *testing.T) (*encryption.ElGamalKeypair, zkencryption.AeKey) {
	t.Helper()
	kp := zktest.GenKeyPair(t)
	aesKey, err := encryption.NewAeKey()
	if err != nil {
		t.Fatal(err)
	}
	return kp, aesKey
}

// encryptBalance encrypts balance under both the ElGamal pubkey and the AE key.
func encryptBalance(t *testing.T, kp *encryption.ElGamalKeypair, aesKey zkencryption.AeKey, balance uint64) (encryption.ElGamalCiphertext, encryption.AeCiphertext) {
	t.Helper()
	balanceCt, err := kp.Pubkey.Encrypt(balance)
	if err != nil {
		t.Fatal(err)
	}
	decryptable, err := encryption.AeEncrypt(aesKey, balance)
	if err != nil {
		t.Fatal(err)
	}
	return balanceCt, decryptable
}

// verifyAll checks every proof in the set, naming the one that fails.
func verifyAll(t *testing.T, proofs map[string]proofdata.ProofData) {
	t.Helper()
	for name, proof := range proofs {
		if err := proof.Verify(); err != nil {
			t.Fatalf("%s proof rejected: %v", name, err)
		}
	}
}

// loHiCiphertext is a grouped ciphertext one half of a lo/hi pair is drawn
// from; both the 2- and 3-handle forms qualify.
type loHiCiphertext interface {
	ToElGamalCiphertext(index int) (encryption.ElGamalCiphertext, error)
}

// handleCiphertexts extracts the pair of ElGamal ciphertexts the key at index
// holds across a lo/hi grouped ciphertext pair.
func handleCiphertexts(t *testing.T, lo, hi loHiCiphertext, index int) (encryption.ElGamalCiphertext, encryption.ElGamalCiphertext) {
	t.Helper()
	loCt, err := lo.ToElGamalCiphertext(index)
	if err != nil {
		t.Fatal(err)
	}
	hiCt, err := hi.ToElGamalCiphertext(index)
	if err != nil {
		t.Fatal(err)
	}
	return loCt, hiCt
}

// decryptHandle recovers the amount the key at index encrypts, recombining its
// lo and hi handles as lo + hi<<loBits the way a recipient or auditor would.
func decryptHandle(t *testing.T, kp *encryption.ElGamalKeypair, lo, hi loHiCiphertext, index int, loBits uint8) uint64 {
	t.Helper()
	loCt, hiCt := handleCiphertexts(t, lo, hi, index)
	return decryptLoHiPair(t, kp, loCt, hiCt, loBits)
}

// decryptLoHiPair recovers the amount an already-extracted lo/hi ElGamal
// ciphertext pair encrypts under kp, recombined as lo + hi<<loBits.
func decryptLoHiPair(t *testing.T, kp *encryption.ElGamalKeypair, loCt, hiCt encryption.ElGamalCiphertext, loBits uint8) uint64 {
	t.Helper()
	loAmount, err := kp.DecryptU32(loCt)
	if err != nil {
		t.Fatal(err)
	}
	hiAmount, err := kp.DecryptU32(hiCt)
	if err != nil {
		t.Fatal(err)
	}
	return loAmount + hiAmount<<loBits
}

// applyToBalance recombines the key's own lo/hi handles into a ciphertext of
// the full amount and applies it to current with op, the way the token program
// recomputes the balance or supply on-chain.
func applyToBalance(
	t *testing.T,
	current encryption.ElGamalCiphertext,
	v CiphertextValidityProofWithAuditorCiphertext,
	index int,
	op func(a, b encryption.ElGamalCiphertext) (encryption.ElGamalCiphertext, error),
) encryption.ElGamalCiphertext {
	t.Helper()
	ctx := v.ProofData.Context
	loCt, hiCt := handleCiphertexts(t, ctx.GroupedCiphertextLo, ctx.GroupedCiphertextHi, index)
	combined, err := encryption.CombineLoHiCiphertexts(loCt, hiCt, AmountLoBitLength)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := op(current, combined)
	if err != nil {
		t.Fatal(err)
	}
	return applied
}

// decryptEquals fails unless kp decrypts ct to want.
func decryptEquals(t *testing.T, kp *encryption.ElGamalKeypair, ct encryption.ElGamalCiphertext, want uint64, what string) {
	t.Helper()
	got, err := kp.DecryptU32(ct)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s decrypts to %d, want %d", what, got, want)
	}
}
