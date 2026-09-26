package token2022

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialMintBurnBurnInstructions burns tokens from a confidential
// balance, appending the verification instructions for the proofs that are in
// sibling instructions.
//
// The instruction fails if the source account is frozen.
func NewConfidentialMintBurnBurnInstructions(
	tokenAccount solana.PublicKey,
	mint solana.PublicKey,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	burnAmountAuditorCiphertextLo encryption.ElGamalCiphertext,
	burnAmountAuditorCiphertextHi encryption.ElGamalCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData],
	ciphertextValidityProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData],
	rangeProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedRangeProofU128Data],
) ([]solana.Instruction, error) {
	accounts := solana.AccountMetaSlice{
		solana.Meta(tokenAccount).WRITE(),
		solana.Meta(mint).WRITE(),
	}
	equalityProofAccount, equalityProofInstructionOffset, err := resolveProofLocation(equalityProofDataLocation)
	if err != nil {
		return nil, err
	}
	ciphertextValidityProofAccount, ciphertextValidityProofInstructionOffset, err := resolveProofLocation(ciphertextValidityProofDataLocation)
	if err != nil {
		return nil, err
	}
	rangeProofAccount, rangeProofInstructionOffset, err := resolveProofLocation(rangeProofDataLocation)
	if err != nil {
		return nil, err
	}
	if equalityProofInstructionOffset != 0 ||
		ciphertextValidityProofInstructionOffset != 0 ||
		rangeProofInstructionOffset != 0 {
		accounts = append(accounts, solana.Meta(solana.SysVarInstructionsPubkey))
	}
	if equalityProofInstructionOffset == 0 {
		accounts = append(accounts, equalityProofAccount)
	}
	if ciphertextValidityProofInstructionOffset == 0 {
		accounts = append(accounts, ciphertextValidityProofAccount)
	}
	if rangeProofInstructionOffset == 0 {
		accounts = append(accounts, rangeProofAccount)
	}

	data := ConfidentialMintBurnBurnData{
		NewDecryptableAvailableBalance:           newDecryptableAvailableBalance,
		BurnAmountAuditorCiphertextLo:            burnAmountAuditorCiphertextLo,
		BurnAmountAuditorCiphertextHi:            burnAmountAuditorCiphertextHi,
		EqualityProofInstructionOffset:           equalityProofInstructionOffset,
		CiphertextValidityProofInstructionOffset: ciphertextValidityProofInstructionOffset,
		RangeProofInstructionOffset:              rangeProofInstructionOffset,
	}

	inner := newConfidentialMintBurnSubInstruction(
		ConfidentialMintBurn_Burn,
		&data,
		accounts,
		authority,
		multisigSigners,
	)
	burnInstruction, err := inner.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	instructions, err := appendVerifyProofInstruction([]solana.Instruction{burnInstruction},
		zkprogram.VerifyCiphertextCommitmentEquality, equalityProofDataLocation)
	if err != nil {
		return nil, err
	}
	instructions, err = appendVerifyProofInstruction(instructions,
		zkprogram.VerifyBatchedGroupedCiphertext3HandlesValidity, ciphertextValidityProofDataLocation)
	if err != nil {
		return nil, err
	}
	return appendVerifyProofInstruction(instructions,
		zkprogram.VerifyBatchedRangeProofU128, rangeProofDataLocation)
}

// ConfidentialMintBurnBurnData is the instruction data for ConfidentialMintBurn_Burn.
type ConfidentialMintBurnBurnData struct {
	// NewDecryptableAvailableBalance is the new decryptable balance of the burner if the burn succeeds.
	NewDecryptableAvailableBalance encryption.AeCiphertext
	// BurnAmountAuditorCiphertextLo is the low bits of the burn amount
	// encrypted under the auditor ElGamal public key.
	BurnAmountAuditorCiphertextLo encryption.ElGamalCiphertext
	// BurnAmountAuditorCiphertextHi is the high bits of the burn amount
	// encrypted under the auditor ElGamal public key.
	BurnAmountAuditorCiphertextHi encryption.ElGamalCiphertext
	// EqualityProofInstructionOffset locates the VerifyCiphertextCommitmentEquality instruction
	// relative to the burn instruction; zero means a context state account.
	EqualityProofInstructionOffset int8
	// CiphertextValidityProofInstructionOffset locates the
	// VerifyBatchedGroupedCiphertext3HandlesValidity instruction relative to
	// the Burn instruction; zero means a context state account.
	CiphertextValidityProofInstructionOffset int8
	// RangeProofInstructionOffset locates the VerifyBatchedRangeProofU128
	// instruction relative to the Burn instruction; zero means a context state account.
	RangeProofInstructionOffset int8
}

const confidentialMintBurnBurnDataSize = aeCiphertextSize + 2*elGamalCiphertextSize + 3*proofOffsetSize

func (d ConfidentialMintBurnBurnData) bytes() []byte {
	out := make([]byte, 0, confidentialMintBurnBurnDataSize)
	out = append(out, d.NewDecryptableAvailableBalance[:]...)
	out = append(out, d.BurnAmountAuditorCiphertextLo[:]...)
	out = append(out, d.BurnAmountAuditorCiphertextHi[:]...)
	out = append(out, byte(d.EqualityProofInstructionOffset))
	out = append(out, byte(d.CiphertextValidityProofInstructionOffset))
	out = append(out, byte(d.RangeProofInstructionOffset))
	return out
}

func (d ConfidentialMintBurnBurnData) MarshalBinary() ([]byte, error) { return d.bytes(), nil }

func (d *ConfidentialMintBurnBurnData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialMintBurnBurnDataSize {
		return fmt.Errorf("token2022: ConfidentialMintBurn Burn data is %d bytes, want %d", len(b), confidentialMintBurnBurnDataSize)
	}
	copy(d.NewDecryptableAvailableBalance[:], b[:36])
	copy(d.BurnAmountAuditorCiphertextLo[:], b[36:100])
	copy(d.BurnAmountAuditorCiphertextHi[:], b[100:164])
	d.EqualityProofInstructionOffset = int8(b[164])
	d.CiphertextValidityProofInstructionOffset = int8(b[165])
	d.RangeProofInstructionOffset = int8(b[166])
	return nil
}
