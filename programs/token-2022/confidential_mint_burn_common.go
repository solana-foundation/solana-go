package token2022

import (
	"github.com/gagliardetto/solana-go"
)

// newConfidentialMintBurnSubInstruction assembles a ConfidentialMintBurn sub-instruction.
func newConfidentialMintBurnSubInstruction(
	subInstruction uint8,
	data ConfidentialMintBurnSubInstructionData,
	accounts solana.AccountMetaSlice,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialMintBurnExtension {
	accounts, signers := confidentialAuthorityAccounts(accounts, authority, multisigSigners)
	return &ConfidentialMintBurnExtension{
		SubInstruction: subInstruction,
		RawData:        data.bytes(),
		Accounts:       accounts,
		Signers:        signers,
	}
}

// cmbNoData is embedded by the data structs of sub-instructions that carry no data.
type cmbNoData struct{ noSubInstructionData }

func (*cmbNoData) UnmarshalBinary(b []byte) error {
	return rejectSubInstructionData(confidentialMintBurnName, b)
}
