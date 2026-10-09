package token2022

import (
	"slices"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token-2022/zkencryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/confidential"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
	"github.com/stretchr/testify/require"
)

// clientProof is a proof a client operation needs: the verification
// instruction, and the context state account the test supplies for it
type clientProof struct {
	verifier zkprogram.ProofInstruction
	account  solana.PublicKey
}

type proofOrder []clientProof

// The proofs of a client operation in the order the program reads them.
// The i-th proof must be in the i-th proof instruction when inlining,
// The i-th context state account must be i-th among the context state accounts otherwise.
var (
	pubkeyValidityProofs     = proofOrder{{zkprogram.VerifyPubkeyValidity, ctContextSingle}}
	zeroCiphertextProofs     = proofOrder{{zkprogram.VerifyZeroCiphertext, ctContextSingle}}
	ciphertextEqualityProofs = proofOrder{{zkprogram.VerifyCiphertextCiphertextEquality, ctContextSingle}}
	withdrawProofs           = proofOrder{
		{zkprogram.VerifyCiphertextCommitmentEquality, ctContextEquality},
		{zkprogram.VerifyBatchedRangeProofU64, ctContextRange},
	}
	splitProofs = proofOrder{
		{zkprogram.VerifyCiphertextCommitmentEquality, ctContextEquality},
		{zkprogram.VerifyBatchedGroupedCiphertext3HandlesValidity, ctContextValidity},
		{zkprogram.VerifyBatchedRangeProofU128, ctContextRange},
	}
	transferWithFeeProofs = proofOrder{
		{zkprogram.VerifyCiphertextCommitmentEquality, ctContextEquality},
		{zkprogram.VerifyBatchedGroupedCiphertext3HandlesValidity, ctContextValidity},
		{zkprogram.VerifyPercentageWithCap, ctContextFeeSigma},
		{zkprogram.VerifyBatchedGroupedCiphertext2HandlesValidity, ctContextFeeValidity},
		{zkprogram.VerifyBatchedRangeProofU256, ctContextRange},
	}
	validityProofAccount = &ProofAccountWithCiphertext{
		ContextStateAccount: ctContextValidity, CiphertextLo: ctCiphertextLo, CiphertextHi: ctCiphertextHi,
	}
)

// The balances and amount of the client operations under test.
const (
	clientBalance = uint64(1_000_000)
	clientSupply  = uint64(5_000_000)
	clientAmount  = uint64(4_321)
)

// forEachProofLocation runs test as parallel subtests with the operation's
// proofs inlined and with them taken from context state accounts.
func forEachProofLocation(t *testing.T, name string, test func(t *testing.T, inline bool)) {
	t.Helper()
	for location, areProofsInline := range map[string]bool{"inline": true, "context": false} {
		t.Run(name+"/"+location, func(t *testing.T) {
			t.Parallel()
			test(t, areProofsInline)
		})
	}
}

// orNil returns v unless the proof it locates is inlined.
func orNil[T any](inline bool, v *T) *T {
	if inline {
		return nil
	}
	return v
}

// clientData decodes the confidential extension instruction a client
// operation leads with into its sub-instruction data, which must be a T.
func clientData[T any](t *testing.T, instructions []solana.Instruction) T {
	t.Helper()
	raw, err := instructions[0].Data()
	require.NoError(t, err)
	decoded, err := DecodeInstruction(instructions[0].Accounts(), raw)
	require.NoError(t, err)
	var data any
	switch extension := decoded.Impl.(type) {
	case *ConfidentialTransferExtension:
		data, err = extension.DecodeSubInstructionData()
	case *ConfidentialTransferFeeExtension:
		data, err = extension.DecodeSubInstructionData()
	case *ConfidentialMintBurnExtension:
		data, err = extension.DecodeSubInstructionData()
	default:
		t.Fatalf("decoded to %T, want a confidential extension instruction", decoded.Impl)
	}
	require.NoError(t, err)
	typed, ok := data.(T)
	if !ok {
		t.Fatalf("sub-instruction decoded to %T, want %T", data, typed)
	}
	return typed
}

// checkClientProofs checks the proofs of a client operation.
func checkClientProofs(
	t *testing.T, instructions []solana.Instruction, proofsAreInline bool, proofs []clientProof, offsets []int8,
) []proofdata.ProofData {
	t.Helper()
	if len(offsets) != len(proofs) {
		t.Fatalf("got %d proof offsets, want one per proof (%d)", len(offsets), len(proofs))
	}
	for i, offset := range offsets {
		want := int8(0)
		if proofsAreInline {
			want = int8(i + 1)
		}
		if offset != want {
			t.Errorf("%s proof offset = %d, want %d", proofs[i].verifier, offset, want)
		}
	}
	if !proofsAreInline {
		if len(instructions) != 1 {
			t.Fatalf("got %d instructions, want 1: no proof is generated for context state accounts", len(instructions))
		}
		// The program takes the context state accounts in proof order, so a
		// client passing one in another proof's slot shows up out of order.
		keys := solana.AccountMetaSlice(instructions[0].Accounts()).GetKeys()
		if keys.Has(solana.SysVarInstructionsPubkey) {
			t.Error("instructions sysvar present, want it only when a proof is inlined")
		}
		previous := -1
		for _, proof := range proofs {
			at := slices.Index(keys, proof.account)
			if at < 0 {
				t.Errorf("context state account %s missing from the instruction accounts", proof.account)
				continue
			}
			if at < previous {
				t.Errorf("%s context state account at index %d, want after the previous proof's at %d",
					proof.verifier, at, previous)
			}
			previous = at
		}
		return nil
	}
	if got, want := len(instructions), 1+len(proofs); got != want {
		t.Fatalf("got %d instructions, want %d", got, want)
	}
	if !solana.AccountMetaSlice(instructions[0].Accounts()).GetKeys().Has(solana.SysVarInstructionsPubkey) {
		t.Error("instructions sysvar missing, want it when a proof is inlined")
	}
	inlinedProofData := make([]proofdata.ProofData, len(proofs))
	for i, proof := range proofs {
		verify := instructions[1+i]
		if got := verify.ProgramID(); !got.Equals(zkprogram.ProgramID) {
			t.Errorf("%s verify program ID = %s, want %s", proof.verifier, got, zkprogram.ProgramID)
		}
		if got := len(verify.Accounts()); got != 0 {
			t.Errorf("%s verify takes %d accounts, want none for an inlined proof", proof.verifier, got)
		}
		data, err := verify.Data()
		require.NoError(t, err)
		if got := zkprogram.ProofInstruction(data[0]); got != proof.verifier {
			t.Errorf("verify instruction %d = %s, want %s", i, got, proof.verifier)
		}
		inlinedProofData[i] = proofdata.NewProofData(proofdata.ProofType(proof.verifier))
		require.NoError(t, inlinedProofData[i].UnmarshalBinary(data[1:]))
		require.NoError(t, inlinedProofData[i].Verify())
	}
	return inlinedProofData
}

// checkCiphertexts checks the ciphertexts of an amount: those
// of validityProofAccount when the proofs are in context state accounts, and
// otherwise an encryption of amount under kp.
func checkCiphertexts(
	t *testing.T, proofsAreInline bool, kp *encryption.ElGamalKeypair, amount uint64, lo, hi encryption.ElGamalCiphertext,
) {
	t.Helper()
	if !proofsAreInline {
		if lo != validityProofAccount.CiphertextLo || hi != validityProofAccount.CiphertextHi {
			t.Error("auditor ciphertexts are not the ones the proof account was given")
		}
		return
	}
	combined, err := encryption.CombineLoHiCiphertexts(lo, hi, confidential.AmountLoBitLength)
	require.NoError(t, err)
	got, err := kp.DecryptU32(combined)
	require.NoError(t, err)
	if got != amount {
		t.Errorf("auditor ciphertexts decrypt to %d, want %d", got, amount)
	}
}

// checkDecryptable fails unless ct decrypts to want under aesKey.
func checkDecryptable(t *testing.T, aesKey zkencryption.AeKey, ct encryption.AeCiphertext, want uint64) {
	t.Helper()
	got, err := encryption.AeDecrypt(aesKey, ct)
	require.NoError(t, err)
	if got != want {
		t.Errorf("decryptable amount decrypts to %d, want %d", got, want)
	}
}

// genKeyPair returns a fresh random ElGamal keypair.
func genKeyPair(t *testing.T) *encryption.ElGamalKeypair {
	t.Helper()
	kp, err := encryption.NewElGamalKeypair()
	if err != nil {
		t.Fatal(err)
	}
	return kp
}

func verifyAll(t *testing.T, proofs map[string]proofdata.ProofData) {
	t.Helper()
	for name, proof := range proofs {
		if err := proof.Verify(); err != nil {
			t.Fatalf("%s proof rejected: %v", name, err)
		}
	}
}
