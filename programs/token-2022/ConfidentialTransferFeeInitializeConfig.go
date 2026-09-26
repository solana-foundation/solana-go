package token2022

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
)

// NewConfidentialTransferFeeInitializeConfigInstruction initializes confidential transfer fees for a mint.
// It must appear in the same transaction as the InitializeMint instruction, otherwise another party can
// initialize the configuration.
//
// A nil authority means none.
func NewConfidentialTransferFeeInitializeConfigInstruction(
	mint solana.PublicKey,
	authority *solana.PublicKey,
	withdrawWithheldAuthorityElGamalPubkey encryption.ElGamalPubkey,
) *ConfidentialTransferFeeExtension {
	data := ConfidentialTransferFeeInitializeConfigData{
		WithdrawWithheldAuthorityElGamalPubkey: withdrawWithheldAuthorityElGamalPubkey,
	}
	if authority != nil {
		data.Authority = *authority
	}
	return &ConfidentialTransferFeeExtension{
		SubInstruction: ConfidentialTransferFee_InitializeConfidentialTransferFeeConfig,
		RawData:        data.bytes(),
		Accounts:       solana.AccountMetaSlice{solana.Meta(mint).WRITE()},
		Signers:        make(solana.AccountMetaSlice, 0),
	}
}

// ConfidentialTransferFeeInitializeConfigData is the instruction data for
// ConfidentialTransferFee_InitializeConfidentialTransferFeeConfig.
type ConfidentialTransferFeeInitializeConfigData struct {
	// Authority is the confidential transfer fee authority. The zero value
	// means no authority.
	Authority solana.PublicKey
	// WithdrawWithheldAuthorityElGamalPubkey is the ElGamal public key used to
	// encrypt withheld fees.
	WithdrawWithheldAuthorityElGamalPubkey encryption.ElGamalPubkey
}

const confidentialTransferFeeInitializeConfigDataSize = pubkeySize + elGamalPubkeySize

func (d ConfidentialTransferFeeInitializeConfigData) bytes() []byte {
	out := make([]byte, 0, confidentialTransferFeeInitializeConfigDataSize)
	out = append(out, d.Authority[:]...)
	out = append(out, d.WithdrawWithheldAuthorityElGamalPubkey[:]...)
	return out
}

func (d ConfidentialTransferFeeInitializeConfigData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialTransferFeeInitializeConfigData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferFeeInitializeConfigDataSize {
		return fmt.Errorf("token2022: ConfidentialTransferFee InitializeConfidentialTransferFeeConfig data is %d bytes, want %d", len(b), confidentialTransferFeeInitializeConfigDataSize)
	}
	copy(d.Authority[:], b[:pubkeySize])
	copy(d.WithdrawWithheldAuthorityElGamalPubkey[:], b[pubkeySize:])
	return nil
}
