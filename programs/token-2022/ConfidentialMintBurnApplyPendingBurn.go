package token2022

import (
	"github.com/gagliardetto/solana-go"
)

// NewConfidentialMintBurnApplyPendingBurnInstruction applies the pending burn amount to the confidential supply.
func NewConfidentialMintBurnApplyPendingBurnInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialMintBurnExtension {
	return newConfidentialMintBurnSubInstruction(
		ConfidentialMintBurn_ApplyPendingBurn,
		&ConfidentialMintBurnApplyPendingBurnData{},
		solana.AccountMetaSlice{solana.Meta(mint).WRITE()},
		authority,
		multisigSigners,
	)
}

// ConfidentialMintBurnApplyPendingBurnData is the instruction data for ConfidentialMintBurn_ApplyPendingBurn, which carries no data.
type ConfidentialMintBurnApplyPendingBurnData struct{ cmbNoData }
