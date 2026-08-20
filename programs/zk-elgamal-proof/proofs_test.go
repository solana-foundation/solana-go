package zk

import (
	"bytes"
	"testing"
)

func TestPubkeyValidityProof(t *testing.T) {
	kp := genKeyPair(t)
	proof, err := PubkeyValidityProof(kp)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}

	tampered := &Proof{Type: proof.Type, Data: bytes.Clone(proof.Data)}
	tampered.Data[len(tampered.Data)-1] ^= 0x01
	err = tampered.Verify()
	if err == nil {
		t.Fatal("tampered proof verified")
	}
	expectStatusError(t, err, PROOF_VERIFICATION_ERROR)
}

func TestZeroCiphertextProof(t *testing.T) {
	kp := genKeyPair(t)
	ct, err := kp.Pubkey.Encrypt(0)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := ZeroCiphertextProof(kp, ct)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}
}

func TestCiphertextCommitmentEqualityProof(t *testing.T) {
	amount := genAmount(t, 1<<32-2) // headroom for the amount+1 mismatch case
	kp := genKeyPair(t)
	ct, err := kp.Pubkey.Encrypt(amount)
	if err != nil {
		t.Fatal(err)
	}
	commitment, opening, err := NewPedersenCommitment(amount)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := CiphertextCommitmentEqualityProof(kp, ct, commitment, opening, amount)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}

	// Inconsistent input must be rejected at generation time.
	wrongCt, err := kp.Pubkey.Encrypt(amount + 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = CiphertextCommitmentEqualityProof(kp, wrongCt, commitment, opening, amount)
	expectStatusError(t, err, PROOF_GENERATION_ERROR)
}

func TestCiphertextCiphertextEqualityProof(t *testing.T) {
	amount := genAmount(t, 1<<32-1)
	source := genKeyPair(t)
	dest := genKeyPair(t)

	sourceCt, err := source.Pubkey.Encrypt(amount)
	if err != nil {
		t.Fatal(err)
	}
	destOpening, err := NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	destCt, err := dest.Pubkey.EncryptWith(amount, destOpening)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := CiphertextCiphertextEqualityProof(source, dest.Pubkey, sourceCt, destCt, destOpening, amount)
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
	deltaAmount := genAmount(t, 1<<32-1)
	percentageCommitment, percentageOpening, err := NewPedersenCommitment(percentageAmount)
	if err != nil {
		t.Fatal(err)
	}
	deltaCommitment, deltaOpening, err := NewPedersenCommitment(deltaAmount)
	if err != nil {
		t.Fatal(err)
	}
	claimedCommitment, claimedOpening, err := NewPedersenCommitment(deltaAmount)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := PercentageWithCapProof(
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
	amounts := []uint64{genAmount(t, 1<<32-1), genAmount(t, 1<<32-1)}
	bitLengths := []uint8{32, 32}
	commitments := make([]PedersenCommitment, len(amounts))
	openings := make([]PedersenOpening, len(amounts))
	for i, amount := range amounts {
		var err error
		commitments[i], openings[i], err = NewPedersenCommitment(amount)
		if err != nil {
			t.Fatal(err)
		}
	}
	proof, err := BatchedRangeProofU64(commitments, amounts, bitLengths, openings)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}

	// Bit lengths that do not sum to 64 are rejected before reaching the
	// prover.
	if _, err := BatchedRangeProofU64(commitments, amounts, []uint8{32, 16}, openings); err == nil {
		t.Fatal("bit lengths summing to 48 accepted for a u64 range proof")
	}

	// An amount inconsistent with its commitment yields a proof that fails
	// verification (the builder does not validate consistency up front).
	inconsistent := []uint64{amounts[0] + 1, amounts[1]}
	badProof, err := BatchedRangeProofU64(commitments, inconsistent, bitLengths, openings)
	if err == nil {
		if err := badProof.Verify(); err == nil {
			t.Fatal("proof over inconsistent amounts verified")
		}
	}
}

func TestBatchedRangeProofU256(t *testing.T) {
	amounts := []uint64{
		genAmount(t, 1<<63), genAmount(t, 1<<63),
		genAmount(t, 1<<63), genAmount(t, 1<<63),
	}
	bitLengths := []uint8{64, 64, 64, 64}
	commitments := make([]PedersenCommitment, len(amounts))
	openings := make([]PedersenOpening, len(amounts))
	for i, amount := range amounts {
		var err error
		commitments[i], openings[i], err = NewPedersenCommitment(amount)
		if err != nil {
			t.Fatal(err)
		}
	}
	proof, err := BatchedRangeProofU256(commitments, amounts, bitLengths, openings)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}
}

func TestGroupedCiphertextValidityProofs(t *testing.T) {
	amount := genAmount(t, 1<<32-1)
	kp2 := [2]*ElGamalKeypair{genKeyPair(t), genKeyPair(t)}
	pubkeys2 := [2]ElGamalPubkey{kp2[0].Pubkey, kp2[1].Pubkey}
	kp3 := [3]*ElGamalKeypair{genKeyPair(t), genKeyPair(t), genKeyPair(t)}
	pubkeys3 := [3]ElGamalPubkey{kp3[0].Pubkey, kp3[1].Pubkey, kp3[2].Pubkey}

	opening, err := NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}

	grouped2, err := GroupedElGamalEncrypt2(pubkeys2, amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	proof2, err := GroupedCiphertext2HandlesValidityProof(pubkeys2, grouped2, amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof2.Verify(); err != nil {
		t.Fatalf("2-handle validity proof rejected: %v", err)
	}

	grouped3, err := GroupedElGamalEncrypt3(pubkeys3, amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	proof3, err := GroupedCiphertext3HandlesValidityProof(pubkeys3, grouped3, amount, opening)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof3.Verify(); err != nil {
		t.Fatalf("3-handle validity proof rejected: %v", err)
	}
}

func TestBatchedGroupedCiphertext2HandlesValidityProof(t *testing.T) {
	amountLo := genAmount(t, 1<<16-1)
	amountHi := genAmount(t, 1<<32-1)
	kps := [2]*ElGamalKeypair{genKeyPair(t), genKeyPair(t)}
	pubkeys := [2]ElGamalPubkey{kps[0].Pubkey, kps[1].Pubkey}

	openingLo, err := NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	openingHi, err := NewPedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	groupedLo, err := GroupedElGamalEncrypt2(pubkeys, amountLo, openingLo)
	if err != nil {
		t.Fatal(err)
	}
	groupedHi, err := GroupedElGamalEncrypt2(pubkeys, amountHi, openingHi)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := BatchedGroupedCiphertext2HandlesValidityProof(
		pubkeys, groupedLo, groupedHi, amountLo, amountHi, openingLo, openingHi)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.Verify(); err != nil {
		t.Fatalf("valid proof rejected: %v", err)
	}
}
