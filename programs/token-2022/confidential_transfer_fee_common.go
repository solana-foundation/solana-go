package token2022

import (
	"github.com/gagliardetto/solana-go"
)

// newConfidentialTransferFeeSubInstruction assembles a ConfidentialTransferFee sub-instruction.
func newConfidentialTransferFeeSubInstruction(
	subInstruction uint8,
	data ConfidentialTransferFeeSubInstructionData,
	accounts solana.AccountMetaSlice,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferFeeExtension {
	accounts, signers := confidentialAuthorityAccounts(accounts, authority, multisigSigners)
	return &ConfidentialTransferFeeExtension{
		SubInstruction: subInstruction,
		RawData:        data.bytes(),
		Accounts:       accounts,
		Signers:        signers,
	}
}

// ctfNoData is embedded by the data structs of sub-instructions that carry no
// data beyond the sub-instruction byte.
type ctfNoData struct{ noSubInstructionData }

func (*ctfNoData) UnmarshalBinary(b []byte) error {
	return rejectSubInstructionData(confidentialTransferFeeName, b)
}
