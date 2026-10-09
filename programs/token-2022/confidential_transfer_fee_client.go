package token2022

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// ConfidentialTransferWithdrawWithheldTokensFromMint withdraws the
// confidential fees withheld in a mint into the available balance of
// destinationAccount.
//
// withheldAmount is the mint's withheld fees, encrypted under the withdraw
// withheld authority's ElGamal pubkey. newDecryptableAvailableBalance is the
// destination's available balance encrypted under its AE key.
func ConfidentialTransferWithdrawWithheldTokensFromMint(
	destinationAccount solana.PublicKey,
	mint solana.PublicKey,
	withdrawWithheldAuthority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	contextStateAccount *solana.PublicKey,
	withheldAmount encryption.ElGamalCiphertext,
	withdrawWithheldAuthorityElgamalKeypair *encryption.ElGamalKeypair,
	destinationElgamalPubkey encryption.ElGamalPubkey,
	newDecryptableAvailableBalance encryption.AeCiphertext,
) ([]solana.Instruction, error) {
	proofData, err := withheldTokensProofData(contextStateAccount, withheldAmount,
		withdrawWithheldAuthorityElgamalKeypair, destinationElgamalPubkey)
	if err != nil {
		return nil, err
	}
	return NewConfidentialTransferFeeWithdrawWithheldTokensFromMintInstructions(
		mint, destinationAccount, newDecryptableAvailableBalance,
		withdrawWithheldAuthority, multisigSigners,
		zkprogram.ConfidentialTransferProofLocation(contextStateAccount, 1, proofData),
	)
}

// ConfidentialTransferWithdrawWithheldTokensFromAccounts directly withdraws the
// confidential fees withheld in the sources' accounts into the available balance
// of destinationAccount.
//
// withheldAmount must be the sum of the sources' withheld fees, as
// encryption.AddCiphertexts aggregates them.
func ConfidentialTransferWithdrawWithheldTokensFromAccounts(
	destinationAccount solana.PublicKey,
	mint solana.PublicKey,
	withdrawWithheldAuthority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	contextStateAccount *solana.PublicKey,
	withheldAmount encryption.ElGamalCiphertext,
	withdrawWithheldAuthorityElgamalKeypair *encryption.ElGamalKeypair,
	destinationElgamalPubkey encryption.ElGamalPubkey,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	sources []solana.PublicKey,
) ([]solana.Instruction, error) {
	proofData, err := withheldTokensProofData(contextStateAccount, withheldAmount,
		withdrawWithheldAuthorityElgamalKeypair, destinationElgamalPubkey)
	if err != nil {
		return nil, err
	}
	return NewConfidentialTransferFeeWithdrawWithheldTokensFromAccountsInstructions(
		mint, destinationAccount, newDecryptableAvailableBalance,
		withdrawWithheldAuthority, multisigSigners, sources,
		zkprogram.ConfidentialTransferProofLocation(contextStateAccount, 1, proofData),
	)
}

// withheldTokensProofData generates the withdraw withheld proof unless it comes from contextStateAccount.
func withheldTokensProofData(
	contextStateAccount *solana.PublicKey,
	withheldAmount encryption.ElGamalCiphertext,
	withdrawWithheldAuthorityElgamalKeypair *encryption.ElGamalKeypair,
	destinationElgamalPubkey encryption.ElGamalPubkey,
) (*proofdata.CiphertextCiphertextEqualityProofData, error) {
	if contextStateAccount != nil {
		return nil, nil
	}
	withheld, err := withdrawWithheldAuthorityElgamalKeypair.DecryptU32(withheldAmount)
	if err != nil {
		return nil, err
	}
	return proveReencryption(withdrawWithheldAuthorityElgamalKeypair, destinationElgamalPubkey, withheldAmount, withheld)
}

// ConfidentialTransferHarvestWithheldTokensToMint moves the confidential fees
// withheld in the sources accounts into the mint.
func ConfidentialTransferHarvestWithheldTokensToMint(
	mint solana.PublicKey,
	sources []solana.PublicKey,
) ([]solana.Instruction, error) {
	built, err := NewConfidentialTransferFeeHarvestWithheldTokensToMintInstruction(mint, sources).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}

// ConfidentialTransferEnableHarvestToMint allows withheld confidential fees to be harvested to the mint.
func ConfidentialTransferEnableHarvestToMint(
	mint solana.PublicKey,
	withdrawWithheldAuthority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) ([]solana.Instruction, error) {
	built, err := NewConfidentialTransferFeeEnableHarvestToMintInstruction(
		mint, withdrawWithheldAuthority, multisigSigners,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}

// ConfidentialTransferDisableHarvestToMint disallows withheld confidential fees to be harvested to the mint.
func ConfidentialTransferDisableHarvestToMint(
	mint solana.PublicKey,
	withdrawWithheldAuthority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) ([]solana.Instruction, error) {
	built, err := NewConfidentialTransferFeeDisableHarvestToMintInstruction(
		mint, withdrawWithheldAuthority, multisigSigners,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}
