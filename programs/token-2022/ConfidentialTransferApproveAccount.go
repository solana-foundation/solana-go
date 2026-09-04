package token2022

import (
	"github.com/gagliardetto/solana-go"
)

// NewConfidentialTransferApproveAccountInstruction approves a token account
// for confidential transfers, on mints where new accounts are not
// auto-approved. authority is the mint's confidential transfer authority and
// must sign directly
func NewConfidentialTransferApproveAccountInstruction(
	accountToApprove solana.PublicKey,
	mint solana.PublicKey,
	authority solana.PublicKey,
) *ConfidentialTransferExtension {
	return newConfidentialTransferInstruction(
		ConfidentialTransfer_ApproveAccount,
		&ConfidentialTransferApproveAccountData{},
		solana.AccountMetaSlice{
			solana.Meta(accountToApprove).WRITE(),
			solana.Meta(mint),
		},
		authority,
		// Multiparty signatures are not supported for this instruction
		nil,
	)
}

// ConfidentialTransferApproveAccountData is the instruction data for
// ConfidentialTransfer_ApproveAccount, which carries no data.
type ConfidentialTransferApproveAccountData struct{ ctNoData }
