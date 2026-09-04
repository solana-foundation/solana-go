package token2022

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialTransferTransferInstructions builds a confidential transfer Transfer instruction.
// Append verification instructions for any associated proofs that are in a sibling intruction.
func NewConfidentialTransferTransferInstructions(
	sourceTokenAccount solana.PublicKey,
	mint solana.PublicKey,
	destinationTokenAccount solana.PublicKey,
	newSourceDecryptableAvailableBalance encryption.AeCiphertext,
	transferAmountAuditorCiphertextLo encryption.ElGamalCiphertext,
	transferAmountAuditorCiphertextHi encryption.ElGamalCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData],
	ciphertextValidityProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData],
	rangeProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedRangeProofU128Data],
) ([]solana.Instruction, error) {
	inner, err := NewConfidentialTransferTransferInstruction(
		sourceTokenAccount, mint, destinationTokenAccount,
		newSourceDecryptableAvailableBalance,
		transferAmountAuditorCiphertextLo, transferAmountAuditorCiphertextHi,
		authority, multisigSigners,
		equalityProofDataLocation, ciphertextValidityProofDataLocation, rangeProofDataLocation,
	)
	if err != nil {
		return nil, err
	}
	built, err := inner.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	instructions := []solana.Instruction{built}
	instructions, err = appendVerifyProofInstruction(instructions,
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

// NewConfidentialTransferTransferInstructions builds a confidential transfer Transfer instruction.
func NewConfidentialTransferTransferInstruction(
	sourceTokenAccount solana.PublicKey,
	mint solana.PublicKey,
	destinationTokenAccount solana.PublicKey,
	newSourceDecryptableAvailableBalance encryption.AeCiphertext,
	transferAmountAuditorCiphertextLo encryption.ElGamalCiphertext,
	transferAmountAuditorCiphertextHi encryption.ElGamalCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData],
	ciphertextValidityProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData],
	rangeProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedRangeProofU128Data],
) (*ConfidentialTransferExtension, error) {
	accounts := solana.AccountMetaSlice{
		solana.Meta(sourceTokenAccount).WRITE(),
		solana.Meta(mint),
		solana.Meta(destinationTokenAccount).WRITE(),
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
	data := ConfidentialTransferTransferData{
		NewSourceDecryptableAvailableBalance:     newSourceDecryptableAvailableBalance,
		TransferAmountAuditorCiphertextLo:        transferAmountAuditorCiphertextLo,
		TransferAmountAuditorCiphertextHi:        transferAmountAuditorCiphertextHi,
		EqualityProofInstructionOffset:           equalityProofInstructionOffset,
		CiphertextValidityProofInstructionOffset: ciphertextValidityProofInstructionOffset,
		RangeProofInstructionOffset:              rangeProofInstructionOffset,
	}
	return newConfidentialTransferInstruction(
		ConfidentialTransfer_Transfer,
		&data,
		accounts,
		authority,
		multisigSigners,
	), nil
}

// ConfidentialTransferTransferData is the instruction data for
// ConfidentialTransfer.Transfer.
type ConfidentialTransferTransferData struct {
	// NewSourceDecryptableAvailableBalance is the new source decryptable
	// balance if the transfer succeeds.
	NewSourceDecryptableAvailableBalance encryption.AeCiphertext
	// TransferAmountAuditorCiphertextLo is the low bits of the transfer amount
	// encrypted under the auditor ElGamal public key.
	TransferAmountAuditorCiphertextLo encryption.ElGamalCiphertext
	// TransferAmountAuditorCiphertextHi is the high bits of the transfer
	// amount encrypted under the auditor ElGamal public key.
	TransferAmountAuditorCiphertextHi encryption.ElGamalCiphertext
	// EqualityProofInstructionOffset locates the VerifyCiphertextCommitmentEquality instruction
	// relative to the transfer instruction; zero means a context state account.
	EqualityProofInstructionOffset int8
	// CiphertextValidityProofInstructionOffset locates the VerifyBatchedGroupedCiphertext3HandlesValidity
	// instruction relative to the transfer instruction; zero means a context state account.
	CiphertextValidityProofInstructionOffset int8
	// RangeProofInstructionOffset locates the VerifyBatchedRangeProofU128
	// instruction relative to the transfer instruction; zero means a context state account.
	RangeProofInstructionOffset int8
}

const confidentialTransferTransferDataSize = aeCiphertextSize + 2*elGamalCiphertextSize + 3*proofOffsetSize

func (d ConfidentialTransferTransferData) MarshalBinary() ([]byte, error) {
	out := make([]byte, 0, confidentialTransferTransferDataSize)
	out = append(out, d.NewSourceDecryptableAvailableBalance[:]...)
	out = append(out, d.TransferAmountAuditorCiphertextLo[:]...)
	out = append(out, d.TransferAmountAuditorCiphertextHi[:]...)
	out = append(out, byte(d.EqualityProofInstructionOffset))
	out = append(out, byte(d.CiphertextValidityProofInstructionOffset))
	out = append(out, byte(d.RangeProofInstructionOffset))
	return out, nil
}

func (d *ConfidentialTransferTransferData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferTransferDataSize {
		return fmt.Errorf("token2022: Transfer data is %d bytes, want %d", len(b), confidentialTransferTransferDataSize)
	}
	copy(d.NewSourceDecryptableAvailableBalance[:], b[:36])
	copy(d.TransferAmountAuditorCiphertextLo[:], b[36:100])
	copy(d.TransferAmountAuditorCiphertextHi[:], b[100:164])
	d.EqualityProofInstructionOffset = int8(b[164])
	d.CiphertextValidityProofInstructionOffset = int8(b[165])
	d.RangeProofInstructionOffset = int8(b[166])
	return nil
}

func (d ConfidentialTransferTransferData) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return ctMarshalData(encoder, d)
}

func (d *ConfidentialTransferTransferData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return ctUnmarshalData(decoder, d)
}
