package proofdata

import (
	"testing"

	zk "github.com/gagliardetto/solana-go/programs/zk-elgamal-proof"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
)

func TestPubkeyValidityProof(t *testing.T) {
	kp := zktest.GenKeyPair(t)
	proof, err := NewPubkeyValidityProofData(kp)
	if err != nil {
		t.Fatal(err)
	}
	if proof.Context.Pubkey != kp.Pubkey {
		t.Fatal("proof context does not echo the proved pubkey")
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}

	tampered := *proof
	tampered.Proof[len(tampered.Proof)-1] ^= 0x01
	err = tampered.Verify()
	if err == nil {
		t.Fatal("tampered proof verified")
	}
	zktest.ExpectStatusError(t, err, zk.PROOF_VERIFICATION_ERROR)
}

func TestZeroCiphertextProof(t *testing.T) {
	kp := zktest.GenKeyPair(t)
	ct, err := kp.Pubkey.Encrypt(0)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := NewZeroCiphertextProofData(kp, ct)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}
}

func TestCiphertextCommitmentEqualityProof(t *testing.T) {
	amount := zktest.GenAmount(t, 1<<32-2) // headroom for the amount+1 mismatch case
	kp := zktest.GenKeyPair(t)
	ct, err := kp.Pubkey.Encrypt(amount)
	if err != nil {
		t.Fatal(err)
	}
	commitment, opening, err := encryption.NewPedersenCommitment(amount)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := NewCiphertextCommitmentEqualityProofData(kp, ct, commitment, opening, amount)
	if err != nil {
		t.Fatal(err)
	}
	if proof.Context.Ciphertext != ct || proof.Context.Commitment != commitment {
		t.Fatal("proof context does not echo the proved ciphertext and commitment")
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}

	// Inconsistent input must be rejected at generation time.
	wrongCt, err := kp.Pubkey.Encrypt(amount + 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewCiphertextCommitmentEqualityProofData(kp, wrongCt, commitment, opening, amount)
	zktest.ExpectStatusError(t, err, zk.PROOF_GENERATION_ERROR)
}

func TestCiphertextCiphertextEqualityProof(t *testing.T) {
	amount := zktest.GenAmount(t, 1<<32-1)
	source := zktest.GenKeyPair(t)
	dest := zktest.GenKeyPair(t)

	sourceCt, err := source.Pubkey.Encrypt(amount)
	if err != nil {
		t.Fatal(err)
	}
	destOpening, err := encryption.NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	destCt, err := dest.Pubkey.EncryptWith(amount, destOpening)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := NewCiphertextCiphertextEqualityProofData(source, dest.Pubkey, sourceCt, destCt, destOpening, amount)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}
}

// Capped branch: the percentage amount equals maxValue, delta and claimed are
// independent commitments to the same delta amount.
func TestPercentageWithCapProof(t *testing.T) {
	const (
		maxValue         = 3
		percentageAmount = 3
	)
	deltaAmount := zktest.GenAmount(t, 1<<32-1)
	percentageCommitment, percentageOpening, err := encryption.NewPedersenCommitment(percentageAmount)
	if err != nil {
		t.Fatal(err)
	}
	deltaCommitment, deltaOpening, err := encryption.NewPedersenCommitment(deltaAmount)
	if err != nil {
		t.Fatal(err)
	}
	claimedCommitment, claimedOpening, err := encryption.NewPedersenCommitment(deltaAmount)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := NewPercentageWithCapProofData(
		percentageCommitment, percentageOpening, percentageAmount,
		deltaCommitment, deltaOpening, deltaAmount,
		claimedCommitment, claimedOpening, maxValue)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}
}

func TestBatchedRangeProofU64(t *testing.T) {
	amounts := []uint64{zktest.GenAmount(t, 1<<32-1), zktest.GenAmount(t, 1<<32-1)}
	bitLengths := []uint8{32, 32}
	commitments := make([]encryption.PedersenCommitment, len(amounts))
	openings := make([]encryption.PedersenOpening, len(amounts))
	for i, amount := range amounts {
		var err error
		commitments[i], openings[i], err = encryption.NewPedersenCommitment(amount)
		if err != nil {
			t.Fatal(err)
		}
	}
	proof, err := NewBatchedRangeProofU64Data(commitments, amounts, bitLengths, openings)
	if err != nil {
		t.Fatal(err)
	}
	if proof.Context.Commitments[0] != commitments[0] || proof.Context.BitLengths[1] != 32 {
		t.Fatal("range proof context does not echo the committed inputs")
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}

	// Bit lengths that do not sum to 64 are rejected before reaching the
	// prover.
	if _, err := NewBatchedRangeProofU64Data(commitments, amounts, []uint8{32, 16}, openings); err == nil {
		t.Fatal("bit lengths summing to 48 accepted for a u64 range proof")
	}

	// An amount inconsistent with its commitment yields a proof that fails
	// verification (the builder does not validate consistency up front).
	inconsistent := []uint64{amounts[0] + 1, amounts[1]}
	badProof, err := NewBatchedRangeProofU64Data(commitments, inconsistent, bitLengths, openings)
	if err == nil {
		if err := badProof.Verify(); err == nil {
			t.Fatal("proof over inconsistent amounts verified")
		}
	}
}

func TestBatchedRangeProofU256(t *testing.T) {
	amounts := []uint64{
		zktest.GenAmount(t, 1<<63), zktest.GenAmount(t, 1<<63),
		zktest.GenAmount(t, 1<<63), zktest.GenAmount(t, 1<<63),
	}
	bitLengths := []uint8{64, 64, 64, 64}
	commitments := make([]encryption.PedersenCommitment, len(amounts))
	openings := make([]encryption.PedersenOpening, len(amounts))
	for i, amount := range amounts {
		var err error
		commitments[i], openings[i], err = encryption.NewPedersenCommitment(amount)
		if err != nil {
			t.Fatal(err)
		}
	}
	proof, err := NewBatchedRangeProofU256Data(commitments, amounts, bitLengths, openings)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}
}

func TestGroupedCiphertextValidityProofs(t *testing.T) {
	amount := zktest.GenAmount(t, 1<<32-1)
	kp2 := [2]*encryption.ElGamalKeypair{zktest.GenKeyPair(t), zktest.GenKeyPair(t)}
	pubkeys2 := [2]encryption.ElGamalPubkey{kp2[0].Pubkey, kp2[1].Pubkey}
	kp3 := [3]*encryption.ElGamalKeypair{zktest.GenKeyPair(t), zktest.GenKeyPair(t), zktest.GenKeyPair(t)}
	pubkeys3 := [3]encryption.ElGamalPubkey{kp3[0].Pubkey, kp3[1].Pubkey, kp3[2].Pubkey}

	opening, err := encryption.NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}

	grouped2, err := encryption.GroupedElGamalEncrypt2(pubkeys2, amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	proof2, err := NewGroupedCiphertext2HandlesValidityProofData(pubkeys2, grouped2, amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof2.Verify(); err != nil {
		t.Fatalf("2-handle validity proof rejected: %v", err)
	}

	grouped3, err := encryption.GroupedElGamalEncrypt3(pubkeys3, amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	proof3, err := NewGroupedCiphertext3HandlesValidityProofData(pubkeys3, grouped3, amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof3.Verify(); err != nil {
		t.Fatalf("3-handle validity proof rejected: %v", err)
	}
}

func TestBatchedGroupedCiphertext2HandlesValidityProof(t *testing.T) {
	amountLo := zktest.GenAmount(t, 1<<16-1)
	amountHi := zktest.GenAmount(t, 1<<32-1)
	kps := [2]*encryption.ElGamalKeypair{zktest.GenKeyPair(t), zktest.GenKeyPair(t)}
	pubkeys := [2]encryption.ElGamalPubkey{kps[0].Pubkey, kps[1].Pubkey}

	openingLo, err := encryption.NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	openingHi, err := encryption.NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	groupedLo, err := encryption.GroupedElGamalEncrypt2(pubkeys, amountLo, openingLo)
	if err != nil {
		t.Fatal(err)
	}
	groupedHi, err := encryption.GroupedElGamalEncrypt2(pubkeys, amountHi, openingHi)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := NewBatchedGroupedCiphertext2HandlesValidityProofData(
		pubkeys, groupedLo, groupedHi, amountLo, amountHi, openingLo, openingHi)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}
}
