package token2022

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialMintBurnRotateSupplyElGamalPubkeyInstructions rotates the
// ElGamal pubkey used to encrypt the confidential supply, appending the
// verification instruction when the proof is in a sibling instruction.
//
// The pending burn amount must be zero for the instruction to succeed.
func NewConfidentialMintBurnRotateSupplyElGamalPubkeyInstructions(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	newSupplyElGamalPubkey encryption.ElGamalPubkey,
	ciphertextEqualityProofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData],
) ([]solana.Instruction, error) {
	proofAccount, proofInstructionOffset, err := resolveProofLocation(ciphertextEqualityProofDataLocation)
	if err != nil {
		return nil, err
	}
	data := ConfidentialMintBurnRotateSupplyElGamalPubkeyData{
		NewSupplyElGamalPubkey: newSupplyElGamalPubkey,
		ProofInstructionOffset: proofInstructionOffset,
	}
	rotatePubkeyInstruction := newConfidentialMintBurnSubInstruction(
		ConfidentialMintBurn_RotateSupplyElGamalPubkey,
		&data,
		solana.AccountMetaSlice{
			solana.Meta(mint).WRITE(),
			proofAccount,
		},
		authority,
		multisigSigners,
	)
	builtRotatePubkeyInstruction, err := rotatePubkeyInstruction.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return appendVerifyProofInstruction([]solana.Instruction{builtRotatePubkeyInstruction},
		zkprogram.VerifyCiphertextCiphertextEquality, ciphertextEqualityProofDataLocation)
}

// ConfidentialMintBurnRotateSupplyElGamalPubkeyData is the instruction data for ConfidentialMintBurn_RotateSupplyElGamalPubkey.
type ConfidentialMintBurnRotateSupplyElGamalPubkeyData struct {
	// NewSupplyElGamalPubkey is the new ElGamal pubkey for supply encryption.
	NewSupplyElGamalPubkey encryption.ElGamalPubkey
	// ProofInstructionOffset locates the VerifyCiphertextCiphertextEquality
	// instruction relative to the RotateSupplyElGamalPubkey instruction; zero
	// means a context state account.
	ProofInstructionOffset int8
}

const confidentialMintBurnRotateSupplyElGamalPubkeyDataSize = elGamalPubkeySize + proofOffsetSize

func (d ConfidentialMintBurnRotateSupplyElGamalPubkeyData) bytes() []byte {
	out := make([]byte, 0, confidentialMintBurnRotateSupplyElGamalPubkeyDataSize)
	out = append(out, d.NewSupplyElGamalPubkey[:]...)
	out = append(out, byte(d.ProofInstructionOffset))
	return out
}

func (d ConfidentialMintBurnRotateSupplyElGamalPubkeyData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialMintBurnRotateSupplyElGamalPubkeyData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialMintBurnRotateSupplyElGamalPubkeyDataSize {
		return fmt.Errorf("token2022: ConfidentialMintBurn RotateSupplyElGamalPubkey data is %d bytes, want %d", len(b), confidentialMintBurnRotateSupplyElGamalPubkeyDataSize)
	}
	copy(d.NewSupplyElGamalPubkey[:], b[:elGamalPubkeySize])
	d.ProofInstructionOffset = int8(b[elGamalPubkeySize])
	return nil
}
