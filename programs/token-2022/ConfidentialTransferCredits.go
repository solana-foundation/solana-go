package token2022

import (
	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

// Create a `EnableConfidentialCreditsInstruction`.
func NewConfidentialTransferEnableConfidentialCreditsInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newBalanceCreditsInstruction(ConfidentialTransfer_EnableConfidentialCredits,
		&ConfidentialTransferEnableConfidentialCreditsData{},
		tokenAccount, authority, multisigSigners)
}

// Create a `DisableConfidentialCreditsInstruction`.
func NewConfidentialTransferDisableConfidentialCreditsInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newBalanceCreditsInstruction(ConfidentialTransfer_DisableConfidentialCredits,
		&ConfidentialTransferDisableConfidentialCreditsData{},
		tokenAccount, authority, multisigSigners)
}

// Create a `EnableNonConfidentialCreditsInstruction`.
func NewConfidentialTransferEnableNonConfidentialCreditsInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newBalanceCreditsInstruction(ConfidentialTransfer_EnableNonConfidentialCredits,
		&ConfidentialTransferEnableNonConfidentialCreditsData{},
		tokenAccount, authority, multisigSigners)
}

// Create a `DisableNonConfidentialCreditsInstruction`.
func NewConfidentialTransferDisableNonConfidentialCreditsInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newBalanceCreditsInstruction(ConfidentialTransfer_DisableNonConfidentialCredits,
		&ConfidentialTransferDisableNonConfidentialCreditsData{},
		tokenAccount, authority, multisigSigners)
}

func newBalanceCreditsInstruction(
	subInstruction uint8,
	data ag_binary.EncoderDecoder,
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newConfidentialTransferInstruction(
		subInstruction,
		data,
		solana.AccountMetaSlice{
			solana.Meta(tokenAccount).WRITE(),
		},
		authority,
		multisigSigners,
	)
}

// The balance credit sub-instructions carry no data.
type (
	ConfidentialTransferEnableConfidentialCreditsData     struct{ ctNoData }
	ConfidentialTransferDisableConfidentialCreditsData    struct{ ctNoData }
	ConfidentialTransferEnableNonConfidentialCreditsData  struct{ ctNoData }
	ConfidentialTransferDisableNonConfidentialCreditsData struct{ ctNoData }
)
