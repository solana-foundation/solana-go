package token2022

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
)

// NewConfidentialTransferUpdateMintInstruction updates the confidential
// transfer mint configuration for a mint.
func NewConfidentialTransferUpdateMintInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	autoApproveNewAccounts bool,
	auditorElGamalPubkey *encryption.ElGamalPubkey,
) *ConfidentialTransferExtension {
	data := ConfidentialTransferUpdateMintData{
		AutoApproveNewAccounts: autoApproveNewAccounts,
	}
	if auditorElGamalPubkey != nil {
		data.AuditorElGamalPubkey = *auditorElGamalPubkey
	}
	return newConfidentialTransferInstruction(
		ConfidentialTransfer_UpdateMint,
		&data,
		solana.AccountMetaSlice{solana.Meta(mint).WRITE()},
		authority,
		multisigSigners,
	)
}

// ConfidentialTransferUpdateMintData is the instruction data for
// ConfidentialTransfer_UpdateMint.
type ConfidentialTransferUpdateMintData struct {
	// AutoApproveNewAccounts determines if newly configured accounts must be
	// approved by the authority before they may be used.
	AutoApproveNewAccounts bool
	// New authority to decode any transfer amount in a confidential transfer.
	// The zero value means no auditor.
	AuditorElGamalPubkey encryption.ElGamalPubkey
}

const confidentialTransferUpdateMintDataSize = boolSize + elGamalPubkeySize

func (d ConfidentialTransferUpdateMintData) MarshalBinary() ([]byte, error) {
	out := make([]byte, 0, confidentialTransferUpdateMintDataSize)
	out = append(out, boolToByte(d.AutoApproveNewAccounts))
	out = append(out, d.AuditorElGamalPubkey[:]...)
	return out, nil
}

func (d *ConfidentialTransferUpdateMintData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferUpdateMintDataSize {
		return fmt.Errorf("token2022: UpdateMint data is %d bytes, want %d", len(b), confidentialTransferUpdateMintDataSize)
	}
	d.AutoApproveNewAccounts = b[0] != 0
	copy(d.AuditorElGamalPubkey[:], b[1:])
	return nil
}

func (d ConfidentialTransferUpdateMintData) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return ctMarshalData(encoder, d)
}

func (d *ConfidentialTransferUpdateMintData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return ctUnmarshalData(decoder, d)
}
