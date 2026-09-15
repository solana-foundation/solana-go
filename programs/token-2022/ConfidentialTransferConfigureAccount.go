package token2022

import (
	"encoding/binary"
	"fmt"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialTransferConfigureAccountInstructions builds a ConfigureAccount instruction.
// When proof location is a sibling instruction (must be offset 1), append its verification instruction.
func NewConfidentialTransferConfigureAccountInstructions(
	tokenAccount solana.PublicKey,
	mint solana.PublicKey,
	decryptableZeroBalance encryption.AeCiphertext,
	maximumPendingBalanceCreditCounter uint64,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.PubkeyValidityProofData],
) ([]solana.Instruction, error) {
	accountConfigInstruction, err := NewConfidentialTransferInnerConfigureAccountInstruction(
		tokenAccount, mint, decryptableZeroBalance, maximumPendingBalanceCreditCounter,
		authority, multisigSigners, proofDataLocation,
	)
	if err != nil {
		return nil, err
	}
	built, err := accountConfigInstruction.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	instructions := []solana.Instruction{built}
	return appendVerifyProofInstruction(instructions,
		zkprogram.VerifyPubkeyValidity, proofDataLocation)
}

// NewConfidentialTransferInnerConfigureAccountInstruction builds a ConfigureAccount instruction.
func NewConfidentialTransferInnerConfigureAccountInstruction(
	tokenAccount solana.PublicKey,
	mint solana.PublicKey,
	decryptableZeroBalance encryption.AeCiphertext,
	maximumPendingBalanceCreditCounter uint64,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.PubkeyValidityProofData],
) (*ConfidentialTransferExtension, error) {
	accounts := solana.AccountMetaSlice{
		solana.Meta(tokenAccount).WRITE(),
		solana.Meta(mint),
	}
	proofLocationAccount, proofInstructionOffset, err := resolveProofLocation(proofDataLocation)
	if err != nil {
		return nil, err
	}
	accounts = append(accounts, proofLocationAccount)

	data := ConfidentialTransferConfigureAccountData{
		DecryptableZeroBalance:             decryptableZeroBalance,
		MaximumPendingBalanceCreditCounter: maximumPendingBalanceCreditCounter,
		ProofInstructionOffset:             proofInstructionOffset,
	}
	return newConfidentialTransferSubInstruction(
		ConfidentialTransfer_ConfigureAccount,
		&data,
		accounts,
		authority,
		multisigSigners,
	), nil
}

// ConfidentialTransferConfigureAccountData is the instruction data for
// ConfidentialTransfer_ConfigureAccount.
type ConfidentialTransferConfigureAccountData struct {
	// DecryptableZeroBalance is the decryptable balance (always 0) once the
	// configure account succeeds.
	DecryptableZeroBalance encryption.AeCiphertext
	// MaximumPendingBalanceCreditCounter is the maximum number of deposits and
	// transfers that an account can receive before ApplyPendingBalance must be
	// executed.
	MaximumPendingBalanceCreditCounter uint64
	// ProofInstructionOffset locates the VerifyPubkeyValidity instruction
	// relative to this one; zero means a context state account.
	ProofInstructionOffset int8
}

const confidentialTransferConfigureAccountDataSize = aeCiphertextSize + u64Size + proofOffsetSize

func (d ConfidentialTransferConfigureAccountData) bytes() []byte {
	out := make([]byte, 0, confidentialTransferConfigureAccountDataSize)
	out = append(out, d.DecryptableZeroBalance[:]...)
	out = binary.LittleEndian.AppendUint64(out, d.MaximumPendingBalanceCreditCounter)
	out = append(out, byte(d.ProofInstructionOffset))
	return out
}

func (d ConfidentialTransferConfigureAccountData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialTransferConfigureAccountData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferConfigureAccountDataSize {
		return fmt.Errorf("token2022: ConfigureAccount data is %d bytes, want %d", len(b), confidentialTransferConfigureAccountDataSize)
	}
	copy(d.DecryptableZeroBalance[:], b[:36])
	d.MaximumPendingBalanceCreditCounter = binary.LittleEndian.Uint64(b[36:44])
	d.ProofInstructionOffset = int8(b[44])
	return nil
}
