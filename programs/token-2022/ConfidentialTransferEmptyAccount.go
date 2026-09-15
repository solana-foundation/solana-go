package token2022

import (
	"fmt"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialTransferEmptyAccountInstructions builds an `EmptyAccount` instruction
// When proof location is a sibling instruction (must be offset 1), append its verification instruction.
func NewConfidentialTransferEmptyAccountInstructions(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.ZeroCiphertextProofData],
) ([]solana.Instruction, error) {
	inner, err := NewConfidentialTransferInnerEmptyAccountInstruction(
		tokenAccount, authority, multisigSigners, proofDataLocation,
	)
	if err != nil {
		return nil, err
	}
	built, err := inner.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	instructions := []solana.Instruction{built}
	return appendVerifyProofInstruction(instructions,
		zkprogram.VerifyZeroCiphertext, proofDataLocation)
}

// NewConfidentialTransferInnerEmptyAccountInstruction empties the confidential
// balances of a token account so it can be closed, proving that the available
// balance ciphertext encrypts zero.
func NewConfidentialTransferInnerEmptyAccountInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.ZeroCiphertextProofData],
) (*ConfidentialTransferExtension, error) {
	accounts := solana.AccountMetaSlice{
		solana.Meta(tokenAccount).WRITE(),
	}
	proofLocationAccount, proofInstructionOffset, err := resolveProofLocation(proofDataLocation)
	if err != nil {
		return nil, err
	}
	accounts = append(accounts, proofLocationAccount)
	data := ConfidentialTransferEmptyAccountData{ProofInstructionOffset: proofInstructionOffset}
	return newConfidentialTransferSubInstruction(
		ConfidentialTransfer_EmptyAccount,
		&data,
		accounts,
		authority,
		multisigSigners,
	), nil
}

// ConfidentialTransferEmptyAccountData is the instruction data for
// ConfidentialTransfer_EmptyAccount.
type ConfidentialTransferEmptyAccountData struct {
	// ProofInstructionOffset locates the VerifyZeroCiphertext instruction
	// relative to this one; zero means a context state account.
	ProofInstructionOffset int8
}

const confidentialTransferEmptyAccountDataSize = proofOffsetSize

func (d ConfidentialTransferEmptyAccountData) bytes() []byte {
	return []byte{byte(d.ProofInstructionOffset)}
}

func (d ConfidentialTransferEmptyAccountData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialTransferEmptyAccountData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferEmptyAccountDataSize {
		return fmt.Errorf("token2022: EmptyAccount data is %d bytes, want %d", len(b), confidentialTransferEmptyAccountDataSize)
	}
	d.ProofInstructionOffset = int8(b[0])
	return nil
}
