package token2022

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialMintBurnMintInstructions mints tokens into a confidential
// balance, appending the verification instructions for the proofs that are in
// sibling instructions.
//
// The instruction fails if the destination account is frozen.
func NewConfidentialMintBurnMintInstructions(
	tokenAccount solana.PublicKey,
	mint solana.PublicKey,
	mintAmountAuditorCiphertextLo encryption.ElGamalCiphertext,
	mintAmountAuditorCiphertextHi encryption.ElGamalCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData],
	ciphertextValidityProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData],
	rangeProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedRangeProofU128Data],
	newDecryptableSupply encryption.AeCiphertext,
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
	data := ConfidentialMintBurnMintData{
		NewDecryptableSupply:                     newDecryptableSupply,
		MintAmountAuditorCiphertextLo:            mintAmountAuditorCiphertextLo,
		MintAmountAuditorCiphertextHi:            mintAmountAuditorCiphertextHi,
		EqualityProofInstructionOffset:           equalityProofInstructionOffset,
		CiphertextValidityProofInstructionOffset: ciphertextValidityProofInstructionOffset,
		RangeProofInstructionOffset:              rangeProofInstructionOffset,
	}
	inner := newConfidentialMintBurnSubInstruction(
		ConfidentialMintBurn_Mint,
		&data,
		accounts,
		authority,
		multisigSigners,
	)
	built, err := inner.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	instructions, err := appendVerifyProofInstruction([]solana.Instruction{built},
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

// ConfidentialMintBurnMintData is the instruction data for ConfidentialMintBurn_Mint.
type ConfidentialMintBurnMintData struct {
	// NewDecryptableSupply is the new decryptable supply if the mint succeeds.
	NewDecryptableSupply encryption.AeCiphertext
	// MintAmountAuditorCiphertextLo is the low bits of the mint amount
	// encrypted under the auditor ElGamal public key.
	MintAmountAuditorCiphertextLo encryption.ElGamalCiphertext
	// MintAmountAuditorCiphertextHi is the high bits of the mint amount
	// encrypted under the auditor ElGamal public key.
	MintAmountAuditorCiphertextHi encryption.ElGamalCiphertext
	// EqualityProofInstructionOffset locates the
	// VerifyCiphertextCommitmentEquality instruction relative to the Mint
	// instruction; zero means a context state account.
	EqualityProofInstructionOffset int8
	// CiphertextValidityProofInstructionOffset locates the
	// VerifyBatchedGroupedCiphertext3HandlesValidity instruction relative to
	// the Mint instruction; zero means a context state account.
	CiphertextValidityProofInstructionOffset int8
	// RangeProofInstructionOffset locates the VerifyBatchedRangeProofU128
	// instruction relative to the Mint instruction; zero means a context state account.
	RangeProofInstructionOffset int8
}

const confidentialMintBurnMintDataSize = aeCiphertextSize + 2*elGamalCiphertextSize + 3*proofOffsetSize

func (d ConfidentialMintBurnMintData) bytes() []byte {
	out := make([]byte, 0, confidentialMintBurnMintDataSize)
	out = append(out, d.NewDecryptableSupply[:]...)
	out = append(out, d.MintAmountAuditorCiphertextLo[:]...)
	out = append(out, d.MintAmountAuditorCiphertextHi[:]...)
	out = append(out, byte(d.EqualityProofInstructionOffset))
	out = append(out, byte(d.CiphertextValidityProofInstructionOffset))
	out = append(out, byte(d.RangeProofInstructionOffset))
	return out
}

func (d ConfidentialMintBurnMintData) MarshalBinary() ([]byte, error) { return d.bytes(), nil }

func (d *ConfidentialMintBurnMintData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialMintBurnMintDataSize {
		return fmt.Errorf("token2022: ConfidentialMintBurn Mint data is %d bytes, want %d", len(b), confidentialMintBurnMintDataSize)
	}
	copy(d.NewDecryptableSupply[:], b[:36])
	copy(d.MintAmountAuditorCiphertextLo[:], b[36:100])
	copy(d.MintAmountAuditorCiphertextHi[:], b[100:164])
	d.EqualityProofInstructionOffset = int8(b[164])
	d.CiphertextValidityProofInstructionOffset = int8(b[165])
	d.RangeProofInstructionOffset = int8(b[166])
	return nil
}
