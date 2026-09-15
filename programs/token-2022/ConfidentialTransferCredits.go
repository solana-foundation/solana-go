package token2022

import (
	"github.com/solana-foundation/solana-go/v2"
)

// NewConfidentialTransferEnableConfidentialCreditsInstruction creates an EnableConfidentialCredits instruction.
func NewConfidentialTransferEnableConfidentialCreditsInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newCreditsInstruction(ConfidentialTransfer_EnableConfidentialCredits,
		&ConfidentialTransferEnableConfidentialCreditsData{},
		tokenAccount, authority, multisigSigners)
}

// NewConfidentialTransferDisableConfidentialCreditsInstruction creates a DisableConfidentialCredits instruction.
func NewConfidentialTransferDisableConfidentialCreditsInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newCreditsInstruction(ConfidentialTransfer_DisableConfidentialCredits,
		&ConfidentialTransferDisableConfidentialCreditsData{},
		tokenAccount, authority, multisigSigners)
}

// NewConfidentialTransferEnableNonConfidentialCreditsInstruction creates an EnableNonConfidentialCredits instruction.
func NewConfidentialTransferEnableNonConfidentialCreditsInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newCreditsInstruction(ConfidentialTransfer_EnableNonConfidentialCredits,
		&ConfidentialTransferEnableNonConfidentialCreditsData{},
		tokenAccount, authority, multisigSigners)
}

// NewConfidentialTransferDisableNonConfidentialCreditsInstruction creates a DisableNonConfidentialCredits instruction.
func NewConfidentialTransferDisableNonConfidentialCreditsInstruction(
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newCreditsInstruction(ConfidentialTransfer_DisableNonConfidentialCredits,
		&ConfidentialTransferDisableNonConfidentialCreditsData{},
		tokenAccount, authority, multisigSigners)
}

func newCreditsInstruction(
	subInstruction uint8,
	data ConfidentialTransferSubInstructionData,
	tokenAccount solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	return newConfidentialTransferSubInstruction(
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
