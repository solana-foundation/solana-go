package token2022

import (
	"github.com/gagliardetto/solana-go"
)

// NewConfidentialTransferFeeHarvestWithheldTokensToMintInstruction transfers
// the withheld confidential tokens of the source accounts to the mint. It is
// permissionless and succeeds for frozen accounts.
func NewConfidentialTransferFeeHarvestWithheldTokensToMintInstruction(
	mint solana.PublicKey,
	sources []solana.PublicKey,
) *ConfidentialTransferFeeExtension {
	accounts := solana.AccountMetaSlice{solana.Meta(mint).WRITE()}
	for _, source := range sources {
		accounts = append(accounts, solana.Meta(source).WRITE())
	}
	return &ConfidentialTransferFeeExtension{
		SubInstruction: ConfidentialTransferFee_HarvestWithheldTokensToMint,
		RawData:        ConfidentialTransferFeeHarvestWithheldTokensToMintData{}.bytes(),
		Accounts:       accounts,
		Signers:        make(solana.AccountMetaSlice, 0),
	}
}

// NewConfidentialTransferFeeEnableHarvestToMintInstruction configures a mint to accept harvested confidential fees.
func NewConfidentialTransferFeeEnableHarvestToMintInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferFeeExtension {
	return newHarvestToMintInstruction(ConfidentialTransferFee_EnableHarvestToMint,
		&ConfidentialTransferFeeEnableHarvestToMintData{}, mint, authority, multisigSigners)
}

// NewConfidentialTransferFeeDisableHarvestToMintInstruction configures a mint to reject harvested confidential fees.
func NewConfidentialTransferFeeDisableHarvestToMintInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferFeeExtension {
	return newHarvestToMintInstruction(ConfidentialTransferFee_DisableHarvestToMint,
		&ConfidentialTransferFeeDisableHarvestToMintData{}, mint, authority, multisigSigners)
}

func newHarvestToMintInstruction(
	subInstruction uint8,
	data ConfidentialTransferFeeSubInstructionData,
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferFeeExtension {
	return newConfidentialTransferFeeSubInstruction(
		subInstruction,
		data,
		solana.AccountMetaSlice{solana.Meta(mint).WRITE()},
		authority,
		multisigSigners,
	)
}

// The harvest sub-instructions carry no data.
type (
	ConfidentialTransferFeeHarvestWithheldTokensToMintData struct{ ctfNoData }
	ConfidentialTransferFeeEnableHarvestToMintData         struct{ ctfNoData }
	ConfidentialTransferFeeDisableHarvestToMintData        struct{ ctfNoData }
)
