package token2022

import (
	"encoding/binary"
	"fmt"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
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
	return newConfidentialTransferSubInstruction(
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
// ConfidentialTransfer_ApplyPendingBalance.
type ConfidentialTransferApplyPendingBalanceData struct {
	// ExpectedPendingBalanceCreditCounter is the expected number of pending
	// balance credits since the last successful ApplyPendingBalance.
	ExpectedPendingBalanceCreditCounter uint64
	// NewDecryptableAvailableBalance is the new decryptable balance if the
	// pending balance is applied successfully.
	NewDecryptableAvailableBalance encryption.AeCiphertext
}

const confidentialTransferApplyPendingBalanceDataSize = u64Size + aeCiphertextSize

func (d ConfidentialTransferApplyPendingBalanceData) bytes() []byte {
	out := make([]byte, 0, confidentialTransferApplyPendingBalanceDataSize)
	out = binary.LittleEndian.AppendUint64(out, d.ExpectedPendingBalanceCreditCounter)
	out = append(out, d.NewDecryptableAvailableBalance[:]...)
	return out
}

func (d ConfidentialTransferApplyPendingBalanceData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialTransferApplyPendingBalanceData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferApplyPendingBalanceDataSize {
		return fmt.Errorf("token2022: ApplyPendingBalance data is %d bytes, want %d", len(b), confidentialTransferApplyPendingBalanceDataSize)
	}
	d.ExpectedPendingBalanceCreditCounter = binary.LittleEndian.Uint64(b[:8])
	copy(d.NewDecryptableAvailableBalance[:], b[8:])
	return nil
}
