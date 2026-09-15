package token2022

import (
	"encoding/binary"
	"fmt"

	"github.com/solana-foundation/solana-go/v2"
)

// NewConfidentialTransferDepositInstruction deposits tokens from the
// non-confidential balance of a token account into its pending confidential
// balance.
func NewConfidentialTransferDepositInstruction(
	tokenAccount solana.PublicKey,
	mint solana.PublicKey,
	amount uint64,
	decimals uint8,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	data := ConfidentialTransferDepositData{Amount: amount, Decimals: decimals}
	return newConfidentialTransferSubInstruction(
		ConfidentialTransfer_Deposit,
		&data,
		solana.AccountMetaSlice{
			solana.Meta(tokenAccount).WRITE(),
			solana.Meta(mint),
		},
		authority,
		multisigSigners,
	)
}

// ConfidentialTransferDepositData is the instruction data for ConfidentialTransfer_Deposit.
type ConfidentialTransferDepositData struct {
	// Amount is the amount of tokens to deposit.
	Amount uint64
	/// Decimals is the expected number of base 10 digits to the right of the decimal place
	Decimals uint8
}

const confidentialTransferDepositDataSize = u64Size + u8Size

func (d ConfidentialTransferDepositData) bytes() []byte {
	out := make([]byte, 0, confidentialTransferDepositDataSize)
	out = binary.LittleEndian.AppendUint64(out, d.Amount)
	out = append(out, d.Decimals)
	return out
}

func (d ConfidentialTransferDepositData) MarshalBinary() ([]byte, error) { return d.bytes(), nil }

func (d *ConfidentialTransferDepositData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferDepositDataSize {
		return fmt.Errorf("token2022: Deposit data is %d bytes, want %d", len(b), confidentialTransferDepositDataSize)
	}
	d.Amount = binary.LittleEndian.Uint64(b[:8])
	d.Decimals = b[8]
	return nil
}
