package token2022

import (
	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

// Create a `ConfigureAccountWithRegistry` instruction.
func NewConfidentialTransferConfigureAccountWithRegistryInstruction(
	tokenAccount solana.PublicKey,
	mint solana.PublicKey,
	elgamalRegistryAccount solana.PublicKey,
	payer *solana.PublicKey,
) *ConfidentialTransferExtension {
	accounts := solana.AccountMetaSlice{
		solana.Meta(tokenAccount).WRITE(),
		solana.Meta(mint),
		solana.Meta(elgamalRegistryAccount),
	}
	if payer != nil {
		accounts = append(accounts,
			solana.Meta(*payer).WRITE().SIGNER(),
			solana.Meta(solana.SystemProgramID),
		)
	}
	return &ConfidentialTransferExtension{
		BaseVariant: ag_binary.BaseVariant{
			TypeID: ag_binary.TypeIDFromUint8(ConfidentialTransfer_ConfigureAccountWithRegistry),
			Impl:   &ConfidentialTransferConfigureAccountWithRegistryData{},
		},
		Accounts: accounts,
		Signers:  make(solana.AccountMetaSlice, 0),
	}
}

// ConfidentialTransferConfigureAccountWithRegistryData is the instruction data
// for ConfidentialTransfer_ConfigureAccountWithRegistry, which carries no data.
type ConfidentialTransferConfigureAccountWithRegistryData struct{ ctNoData }
