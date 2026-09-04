package token2022

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialTransferEmptyAccountInstructions builds an `EmptyAccount` instruction
// When proof location is a sibling instruction (must be offset 1), append its verification instruction.
func NewConfidentialTransferEmptyAccountInstructions(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.ZeroCiphertextProofData],
) ([]solana.Instruction, error) {
	inner, err := NewConfidentialTransferEmptyAccountInstruction(
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

// NewConfidentialTransferEmptyAccountInstruction empties the confidential
// balances of a token account so it can be closed, proving that the available
// balance ciphertext encrypts zero.
func NewConfidentialTransferEmptyAccountInstruction(
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
	return newConfidentialTransferInstruction(
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

func (d ConfidentialTransferEmptyAccountData) MarshalBinary() ([]byte, error) {
	return []byte{byte(d.ProofInstructionOffset)}, nil
}

func (d *ConfidentialTransferEmptyAccountData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferEmptyAccountDataSize {
		return fmt.Errorf("token2022: EmptyAccount data is %d bytes, want %d", len(b), confidentialTransferEmptyAccountDataSize)
	}
	d.ProofInstructionOffset = int8(b[0])
	return nil
}

func (d ConfidentialTransferEmptyAccountData) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return ctMarshalData(encoder, d)
}

func (d *ConfidentialTransferEmptyAccountData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return ctUnmarshalData(decoder, d)
}
