package token2022

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialTransferWithdrawInstructions builds a confidential transfer Withdraw instruction.
// Append verification instructions for any associated proofs that are in a sibling intruction.
func NewConfidentialTransferWithdrawInstructions(
	tokenAccount solana.PublicKey,
	mint solana.PublicKey,
	amount uint64,
	decimals uint8,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData],
	rangeProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedRangeProofU64Data],
) ([]solana.Instruction, error) {
	inner, err := NewConfidentialTransferWithdrawInstruction(
		tokenAccount, mint, amount, decimals, newDecryptableAvailableBalance,
		authority, multisigSigners, equalityProofDataLocation, rangeProofDataLocation,
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
	return appendVerifyProofInstruction(instructions,
		zkprogram.VerifyBatchedRangeProofU64, rangeProofDataLocation)
}

// NewConfidentialTransferWithdrawInstructions builds a confidential transfer Withdraw instruction.
func NewConfidentialTransferWithdrawInstruction(
	tokenAccount solana.PublicKey,
	mint solana.PublicKey,
	amount uint64,
	decimals uint8,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCommitmentEqualityProofData],
	rangeProofDataLocation zkprogram.ProofLocation[*proofdata.BatchedRangeProofU64Data],
) (*ConfidentialTransferExtension, error) {
	accounts := solana.AccountMetaSlice{
		solana.Meta(tokenAccount).WRITE(),
		solana.Meta(mint),
	}
	equalityProofAccount, equalityProofInstructionOffset, err := resolveProofLocation(equalityProofDataLocation)
	if err != nil {
		return nil, err
	}
	rangeProofAccount, rangeProofInstructionOffset, err := resolveProofLocation(rangeProofDataLocation)
	if err != nil {
		return nil, err
	}
	if equalityProofInstructionOffset != 0 || rangeProofInstructionOffset != 0 {
		accounts = append(accounts, solana.Meta(solana.SysVarInstructionsPubkey))
	}
	if equalityProofInstructionOffset == 0 {
		accounts = append(accounts, equalityProofAccount)
	}
	if rangeProofInstructionOffset == 0 {
		accounts = append(accounts, rangeProofAccount)
	}
	data := ConfidentialTransferWithdrawData{
		Amount:                         amount,
		Decimals:                       decimals,
		NewDecryptableAvailableBalance: newDecryptableAvailableBalance,
		EqualityProofInstructionOffset: equalityProofInstructionOffset,
		RangeProofInstructionOffset:    rangeProofInstructionOffset,
	}
	return newConfidentialTransferInstruction(
		ConfidentialTransfer_Withdraw,
		&data,
		accounts,
		authority,
		multisigSigners,
	), nil
}

// ConfidentialTransferWithdrawData is the instruction data for
// ConfidentialTransfer.Withdraw.
type ConfidentialTransferWithdrawData struct {
	// Amount is the amount of tokens to withdraw.
	Amount uint64
	// Decimals is the expected number of base 10 digits to the right of the decimal place.
	Decimals uint8
	// NewDecryptableAvailableBalance is the new decryptable balance if the withdrawal succeeds.
	NewDecryptableAvailableBalance encryption.AeCiphertext
	// EqualityProofInstructionOffset locates the VerifyCiphertextCommitmentEquality instruction
	// relative to the withdrawal instruction; zero means a context state account.
	EqualityProofInstructionOffset int8
	// RangeProofInstructionOffset locates the VerifyBatchedRangeProofU64
	// instruction relative to the withdrawal instruction; zero means a context state account.
	RangeProofInstructionOffset int8
}

const confidentialTransferWithdrawDataSize = u64Size + u8Size + aeCiphertextSize + 2*proofOffsetSize

func (d ConfidentialTransferWithdrawData) MarshalBinary() ([]byte, error) {
	out := make([]byte, 0, confidentialTransferWithdrawDataSize)
	out = binary.LittleEndian.AppendUint64(out, d.Amount)
	out = append(out, d.Decimals)
	out = append(out, d.NewDecryptableAvailableBalance[:]...)
	out = append(out, byte(d.EqualityProofInstructionOffset))
	out = append(out, byte(d.RangeProofInstructionOffset))
	return out, nil
}

func (d *ConfidentialTransferWithdrawData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferWithdrawDataSize {
		return fmt.Errorf("token2022: Withdraw data is %d bytes, want %d", len(b), confidentialTransferWithdrawDataSize)
	}
	d.Amount = binary.LittleEndian.Uint64(b[:8])
	d.Decimals = b[8]
	copy(d.NewDecryptableAvailableBalance[:], b[9:45])
	d.EqualityProofInstructionOffset = int8(b[45])
	d.RangeProofInstructionOffset = int8(b[46])
	return nil
}

func (d ConfidentialTransferWithdrawData) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return ctMarshalData(encoder, d)
}

func (d *ConfidentialTransferWithdrawData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return ctUnmarshalData(decoder, d)
}
