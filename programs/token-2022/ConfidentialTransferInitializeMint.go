package token2022

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
)

// NewConfidentialTransferInitializeMintInstruction initializes confidential
// transfers for a mint. It must appear earlier in the same transaction as the
// InitializeMint instruction.
//
// A nil authority or auditorElGamalPubkey means none.
func NewConfidentialTransferInitializeMintInstruction(
	mint solana.PublicKey,
	authority *solana.PublicKey,
	autoApproveNewAccounts bool,
	auditorElGamalPubkey *encryption.ElGamalPubkey,
) *ConfidentialTransferExtension {
	data := ConfidentialTransferInitializeMintData{
		AutoApproveNewAccounts: autoApproveNewAccounts,
	}
	if authority != nil {
		data.Authority = *authority
	}
	if auditorElGamalPubkey != nil {
		data.AuditorElGamalPubkey = *auditorElGamalPubkey
	}
	return &ConfidentialTransferExtension{
		BaseVariant: ag_binary.BaseVariant{
			TypeID: ag_binary.TypeIDFromUint8(ConfidentialTransfer_InitializeMint),
			Impl:   &data,
		},
		Accounts: solana.AccountMetaSlice{solana.Meta(mint).WRITE()},
		Signers:  make(solana.AccountMetaSlice, 0),
	}
}

// ConfidentialTransferInitializeMintData is the instruction data for ConfidentialTransfer_InitializeMint.
type ConfidentialTransferInitializeMintData struct {
	// Authority to modify the ConfidentialTransferMint configuration and to
	// approve new accounts. The zero value means no authority.
	Authority solana.PublicKey
	/// Determines if newly configured accounts must be approved by the
	/// `Authority` before they may be used by the user.
	AutoApproveNewAccounts bool
	/// New authority to decode any transfer amount in a confidential transfer.
	// The zero value means no auditor.
	AuditorElGamalPubkey encryption.ElGamalPubkey
}

const confidentialTransferInitializeMintDataSize = accountKeySize + boolSize + elGamalPubkeySize

func (d ConfidentialTransferInitializeMintData) MarshalBinary() ([]byte, error) {
	out := make([]byte, 0, confidentialTransferInitializeMintDataSize)
	out = append(out, d.Authority[:]...)
	out = append(out, boolToByte(d.AutoApproveNewAccounts))
	out = append(out, d.AuditorElGamalPubkey[:]...)
	return out, nil
}

func (d *ConfidentialTransferInitializeMintData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferInitializeMintDataSize {
		return fmt.Errorf("token2022: InitializeMint data is %d bytes, want %d", len(b), confidentialTransferInitializeMintDataSize)
	}
	copy(d.Authority[:], b[:32])
	d.AutoApproveNewAccounts = b[32] != 0
	copy(d.AuditorElGamalPubkey[:], b[33:])
	return nil
}

func (d ConfidentialTransferInitializeMintData) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return ctMarshalData(encoder, d)
}

func (d *ConfidentialTransferInitializeMintData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return ctUnmarshalData(decoder, d)
}
