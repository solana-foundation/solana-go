package token2022

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialTransferFeeWithdrawWithheldTokensFromMintInstructions
// transfers all withheld confidential tokens in the mint to an account,
// appending the verification instruction when the proof is in a sibling
// instruction.
//
// The proof must sit directly after the withdraw instruction; for any other
// offset use NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromMintInstruction.
func NewConfidentialTransferFeeWithdrawWithheldTokensFromMintInstructions(
	mint solana.PublicKey,
	destination solana.PublicKey,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData],
) ([]solana.Instruction, error) {
	withdrawWithheldFromMintInstruction, err := NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromMintInstruction(
		mint, destination, newDecryptableAvailableBalance, authority, multisigSigners, proofDataLocation)
	if err != nil {
		return nil, err
	}
	builtWithdrawWithheldFromMintInstruction, err := withdrawWithheldFromMintInstruction.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return appendVerifyProofInstruction([]solana.Instruction{builtWithdrawWithheldFromMintInstruction},
		zkprogram.VerifyCiphertextCiphertextEquality, proofDataLocation)
}

// NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromMintInstruction
// builds the WithdrawWithheldTokensFromMint instruction on its own, for use
// with a cross-program invoke.
func NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromMintInstruction(
	mint solana.PublicKey,
	destination solana.PublicKey,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData],
) (*ConfidentialTransferFeeExtension, error) {
	proofAccount, proofInstructionOffset, err := resolveProofLocation(proofDataLocation)
	if err != nil {
		return nil, err
	}
	proofData := ConfidentialTransferFeeWithdrawWithheldTokensFromMintData{
		ProofInstructionOffset:         proofInstructionOffset,
		NewDecryptableAvailableBalance: newDecryptableAvailableBalance,
	}
	return newConfidentialTransferFeeSubInstruction(
		ConfidentialTransferFee_WithdrawWithheldTokensFromMint,
		&proofData,
		solana.AccountMetaSlice{
			solana.Meta(mint).WRITE(),
			solana.Meta(destination).WRITE(),
			proofAccount,
		},
		authority,
		multisigSigners,
	), nil
}

// ConfidentialTransferFeeWithdrawWithheldTokensFromMintData is the instruction
// data for ConfidentialTransferFee_WithdrawWithheldTokensFromMint.
type ConfidentialTransferFeeWithdrawWithheldTokensFromMintData struct {
	// ProofInstructionOffset locates the VerifyCiphertextCiphertextEquality
	// instruction relative to the WithdrawWithheldTokensFromMint instruction;
	// zero means a context state account.
	ProofInstructionOffset int8
	// NewDecryptableAvailableBalance is the new decryptable balance in the
	// destination token account.
	NewDecryptableAvailableBalance encryption.AeCiphertext
}

const confidentialTransferFeeWithdrawWithheldTokensFromMintDataSize = proofOffsetSize + aeCiphertextSize

func (d ConfidentialTransferFeeWithdrawWithheldTokensFromMintData) bytes() []byte {
	out := make([]byte, 0, confidentialTransferFeeWithdrawWithheldTokensFromMintDataSize)
	out = append(out, byte(d.ProofInstructionOffset))
	out = append(out, d.NewDecryptableAvailableBalance[:]...)
	return out
}

func (d ConfidentialTransferFeeWithdrawWithheldTokensFromMintData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialTransferFeeWithdrawWithheldTokensFromMintData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferFeeWithdrawWithheldTokensFromMintDataSize {
		return fmt.Errorf("token2022: ConfidentialTransferFee WithdrawWithheldTokensFromMint data is %d bytes, want %d", len(b), confidentialTransferFeeWithdrawWithheldTokensFromMintDataSize)
	}
	d.ProofInstructionOffset = int8(b[0])
	copy(d.NewDecryptableAvailableBalance[:], b[1:])
	return nil
}
