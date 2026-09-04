package token2022

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
)

// NewConfidentialTransferApplyPendingBalanceInstruction rolls the pending
// confidential balance of a token account into its available balance.
func NewConfidentialTransferApplyPendingBalanceInstruction(
	tokenAccount solana.PublicKey,
	expectedPendingBalanceCreditCounter uint64,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	data := ConfidentialTransferApplyPendingBalanceData{
		ExpectedPendingBalanceCreditCounter: expectedPendingBalanceCreditCounter,
		NewDecryptableAvailableBalance:      newDecryptableAvailableBalance,
	}
	return newConfidentialTransferInstruction(
		ConfidentialTransfer_ApplyPendingBalance,
		&data,
		solana.AccountMetaSlice{
			solana.Meta(tokenAccount).WRITE(),
		},
		authority,
		multisigSigners,
	)
}

// ConfidentialTransferApplyPendingBalanceData is the instruction data for
// ConfidentialTransfer.ApplyPendingBalance.
type ConfidentialTransferApplyPendingBalanceData struct {
	// ExpectedPendingBalanceCreditCounter is the expected number of pending
	// balance credits since the last successful ApplyPendingBalance.
	ExpectedPendingBalanceCreditCounter uint64
	// NewDecryptableAvailableBalance is the new decryptable balance if the
	// pending balance is applied successfully.
	NewDecryptableAvailableBalance encryption.AeCiphertext
}

const confidentialTransferApplyPendingBalanceDataSize = u64Size + aeCiphertextSize

func (d ConfidentialTransferApplyPendingBalanceData) MarshalBinary() ([]byte, error) {
	out := make([]byte, 0, confidentialTransferApplyPendingBalanceDataSize)
	out = binary.LittleEndian.AppendUint64(out, d.ExpectedPendingBalanceCreditCounter)
	out = append(out, d.NewDecryptableAvailableBalance[:]...)
	return out, nil
}

func (d *ConfidentialTransferApplyPendingBalanceData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferApplyPendingBalanceDataSize {
		return fmt.Errorf("token2022: ApplyPendingBalance data is %d bytes, want %d", len(b), confidentialTransferApplyPendingBalanceDataSize)
	}
	d.ExpectedPendingBalanceCreditCounter = binary.LittleEndian.Uint64(b[:8])
	copy(d.NewDecryptableAvailableBalance[:], b[8:])
	return nil
}

func (d ConfidentialTransferApplyPendingBalanceData) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return ctMarshalData(encoder, d)
}

func (d *ConfidentialTransferApplyPendingBalanceData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return ctUnmarshalData(decoder, d)
}
