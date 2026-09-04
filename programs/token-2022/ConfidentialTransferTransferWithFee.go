package token2022

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialTransferTransferWithFeeInstructions builds a confidential transfer TransferWithFee instruction.
// Append verification instructions for any associated proofs that are in a sibling intruction.
func NewConfidentialTransferTransferWithFeeInstructions(
	sourceTokenAccount solana.PublicKey,
	mint solana.PublicKey,
	destinationTokenAccount solana.PublicKey,
	newSourceDecryptableAvailableBalance encryption.AeCiphertext,
	transferAmountAuditorCiphertextLo encryption.ElGamalCiphertext,
	transferAmountAuditorCiphertextHi encryption.ElGamalCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData],
	transferAmountCiphertextValidityProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData],
	feeSigmaProofDataLocation zkprogram.ProofLocation[*proofdata.PercentageWithCapProofData],
	feeCiphertextValidityProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext2HandlesValidityProofData],
	rangeProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedRangeProofU256Data],
) ([]solana.Instruction, error) {
	inner, err := NewConfidentialTransferTransferWithFeeInstruction(
		sourceTokenAccount, mint, destinationTokenAccount,
		newSourceDecryptableAvailableBalance,
		transferAmountAuditorCiphertextLo, transferAmountAuditorCiphertextHi,
		authority, multisigSigners,
		equalityProofDataLocation, transferAmountCiphertextValidityProofDataLocation,
		feeSigmaProofDataLocation, feeCiphertextValidityProofDataLocation, rangeProofDataLocation,
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
		zkprogram.VerifyBatchedGroupedCiphertext3HandlesValidity, transferAmountCiphertextValidityProofDataLocation)
	if err != nil {
		return nil, err
	}
	instructions, err = appendVerifyProofInstruction(instructions,
		zkprogram.VerifyPercentageWithCap, feeSigmaProofDataLocation)
	if err != nil {
		return nil, err
	}
	instructions, err = appendVerifyProofInstruction(instructions,
		zkprogram.VerifyBatchedGroupedCiphertext2HandlesValidity, feeCiphertextValidityProofDataLocation)
	if err != nil {
		return nil, err
	}
	return appendVerifyProofInstruction(instructions,
		zkprogram.VerifyBatchedRangeProofU256, rangeProofDataLocation)
}

// NewConfidentialTransferTransferWithFeeInstructions builds a confidential transfer TransferWithFee instruction.
func NewConfidentialTransferTransferWithFeeInstruction(
	sourceTokenAccount solana.PublicKey,
	mint solana.PublicKey,
	destinationTokenAccount solana.PublicKey,
	newSourceDecryptableAvailableBalance encryption.AeCiphertext,
	transferAmountAuditorCiphertextLo encryption.ElGamalCiphertext,
	transferAmountAuditorCiphertextHi encryption.ElGamalCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData],
	transferAmountCiphertextValidityProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext3HandlesValidityProofData],
	feeSigmaProofDataLocation zkprogram.ProofLocation[*proofdata.PercentageWithCapProofData],
	feeCiphertextValidityProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedGroupedCiphertext2HandlesValidityProofData],
	rangeProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedRangeProofU256Data],
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
	transferAmountCiphertextValidityProofAccount, transferAmountCiphertextValidityProofInstructionOffset, err := resolveProofLocation(transferAmountCiphertextValidityProofDataLocation)
	if err != nil {
		return nil, err
	}
	feeSigmaProofAccount, feeSigmaProofInstructionOffset, err := resolveProofLocation(feeSigmaProofDataLocation)
	if err != nil {
		return nil, err
	}
	feeCiphertextValidityProofAccount, feeCiphertextValidityProofInstructionOffset, err := resolveProofLocation(feeCiphertextValidityProofDataLocation)
	if err != nil {
		return nil, err
	}
	rangeProofAccount, rangeProofInstructionOffset, err := resolveProofLocation(rangeProofDataLocation)
	if err != nil {
		return nil, err
	}
	if equalityProofInstructionOffset != 0 ||
		transferAmountCiphertextValidityProofInstructionOffset != 0 ||
		feeSigmaProofInstructionOffset != 0 ||
		feeCiphertextValidityProofInstructionOffset != 0 ||
		rangeProofInstructionOffset != 0 {
		accounts = append(accounts, solana.Meta(solana.SysVarInstructionsPubkey))
	}
	if equalityProofInstructionOffset == 0 {
		accounts = append(accounts, equalityProofAccount)
	}
	if transferAmountCiphertextValidityProofInstructionOffset == 0 {
		accounts = append(accounts, transferAmountCiphertextValidityProofAccount)
	}
	if feeSigmaProofInstructionOffset == 0 {
		accounts = append(accounts, feeSigmaProofAccount)
	}
	if feeCiphertextValidityProofInstructionOffset == 0 {
		accounts = append(accounts, feeCiphertextValidityProofAccount)
	}
	if rangeProofInstructionOffset == 0 {
		accounts = append(accounts, rangeProofAccount)
	}
	data := ConfidentialTransferTransferWithFeeData{
		NewSourceDecryptableAvailableBalance:                   newSourceDecryptableAvailableBalance,
		TransferAmountAuditorCiphertextLo:                      transferAmountAuditorCiphertextLo,
		TransferAmountAuditorCiphertextHi:                      transferAmountAuditorCiphertextHi,
		EqualityProofInstructionOffset:                         equalityProofInstructionOffset,
		TransferAmountCiphertextValidityProofInstructionOffset: transferAmountCiphertextValidityProofInstructionOffset,
		FeeSigmaProofInstructionOffset:                         feeSigmaProofInstructionOffset,
		FeeCiphertextValidityProofInstructionOffset:            feeCiphertextValidityProofInstructionOffset,
		RangeProofInstructionOffset:                            rangeProofInstructionOffset,
	}
	return newConfidentialTransferInstruction(
		ConfidentialTransfer_TransferWithFee,
		&data,
		accounts,
		authority,
		multisigSigners,
	), nil
}

// ConfidentialTransferTransferWithFeeData is the instruction data for
// ConfidentialTransfer.TransferWithFee.
type ConfidentialTransferTransferWithFeeData struct {
	// NewSourceDecryptableAvailableBalance is the new source decryptable
	// balance if the transfer succeeds.
	NewSourceDecryptableAvailableBalance encryption.AeCiphertext
	// TransferAmountAuditorCiphertextLo is the low bits of the transfer amount
	// encrypted under the auditor ElGamal public key.
	TransferAmountAuditorCiphertextLo encryption.ElGamalCiphertext
	// TransferAmountAuditorCiphertextHi is the high bits of the transfer
	// amount encrypted under the auditor ElGamal public key.
	TransferAmountAuditorCiphertextHi encryption.ElGamalCiphertext
	// EqualityProofInstructionOffset locates the
	// VerifyCiphertextCommitmentEquality instruction relative to this one;
	// zero means a context state account.
	EqualityProofInstructionOffset int8
	// TransferAmountCiphertextValidityProofInstructionOffset locates the
	// VerifyBatchedGroupedCiphertext3HandlesValidity instruction relative to
	// the TransferWithFee instruction; zero means a context state account.
	TransferAmountCiphertextValidityProofInstructionOffset int8
	// FeeSigmaProofInstructionOffset locates the VerifyPercentageWithCap
	// instruction relative to the TransferWithFee instruction; zero means a context state account.
	FeeSigmaProofInstructionOffset int8
	// FeeCiphertextValidityProofInstructionOffset locates the
	// VerifyBatchedGroupedCiphertext2HandlesValidity instruction relative to
	// the TransferWithFee instruction; zero means a context state account.
	FeeCiphertextValidityProofInstructionOffset int8
	// RangeProofInstructionOffset locates the VerifyBatchedRangeProofU256
	// instruction relative to the TransferWithFee instruction; zero means a context state account.
	RangeProofInstructionOffset int8
}

const confidentialTransferTransferWithFeeDataSize = aeCiphertextSize + 2*elGamalCiphertextSize + 5*proofOffsetSize

func (d ConfidentialTransferTransferWithFeeData) MarshalBinary() ([]byte, error) {
	out := make([]byte, 0, confidentialTransferTransferWithFeeDataSize)
	out = append(out, d.NewSourceDecryptableAvailableBalance[:]...)
	out = append(out, d.TransferAmountAuditorCiphertextLo[:]...)
	out = append(out, d.TransferAmountAuditorCiphertextHi[:]...)
	out = append(out, byte(d.EqualityProofInstructionOffset))
	out = append(out, byte(d.TransferAmountCiphertextValidityProofInstructionOffset))
	out = append(out, byte(d.FeeSigmaProofInstructionOffset))
	out = append(out, byte(d.FeeCiphertextValidityProofInstructionOffset))
	out = append(out, byte(d.RangeProofInstructionOffset))
	return out, nil
}

func (d *ConfidentialTransferTransferWithFeeData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferTransferWithFeeDataSize {
		return fmt.Errorf("token2022: TransferWithFee data is %d bytes, want %d", len(b), confidentialTransferTransferWithFeeDataSize)
	}
	copy(d.NewSourceDecryptableAvailableBalance[:], b[:36])
	copy(d.TransferAmountAuditorCiphertextLo[:], b[36:100])
	copy(d.TransferAmountAuditorCiphertextHi[:], b[100:164])
	d.EqualityProofInstructionOffset = int8(b[164])
	d.TransferAmountCiphertextValidityProofInstructionOffset = int8(b[165])
	d.FeeSigmaProofInstructionOffset = int8(b[166])
	d.FeeCiphertextValidityProofInstructionOffset = int8(b[167])
	d.RangeProofInstructionOffset = int8(b[168])
	return nil
}

func (d ConfidentialTransferTransferWithFeeData) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return ctMarshalData(encoder, d)
}

func (d *ConfidentialTransferTransferWithFeeData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return ctUnmarshalData(decoder, d)
}
