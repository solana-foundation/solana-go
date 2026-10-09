package token2022

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token-2022/zkencryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/confidential"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// ConfidentialTransferRotateSupplyElGamalPubkey re-encrypts the confidential
// supply of a mint under newSupplyElGamalPubkey.
//
// The mint's pending burn must be applied first.
func ConfidentialTransferRotateSupplyElGamalPubkey(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	currentSupplyElgamalKeypair *encryption.ElGamalKeypair,
	newSupplyElgamalPubkey encryption.ElGamalPubkey,
	aesKey zkencryption.AeKey,
	contextStateAccount *solana.PublicKey,
	accountInfo SupplyAccountInfo,
) ([]solana.Instruction, error) {
	var proofData *proofdata.CiphertextCiphertextEqualityProofData
	if contextStateAccount == nil {
		var err error
		if proofData, err = accountInfo.GenerateRotateSupplyElGamalPubkeyProof(
			currentSupplyElgamalKeypair, newSupplyElgamalPubkey, aesKey,
		); err != nil {
			return nil, err
		}
	}
	return NewConfidentialMintBurnRotateSupplyElGamalPubkeyInstructions(
		mint, authority, multisigSigners, newSupplyElgamalPubkey,
		zkprogram.ConfidentialTransferProofLocation(contextStateAccount, 1, proofData),
	)
}

// ConfidentialTransferUpdateDecryptSupply replaces the decryptable supply cache of a mint.
func ConfidentialTransferUpdateDecryptSupply(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	newDecryptableSupply encryption.AeCiphertext,
) ([]solana.Instruction, error) {
	built, err := NewConfidentialMintBurnUpdateDecryptableSupplyInstruction(
		mint, authority, multisigSigners, newDecryptableSupply,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}

// ConfidentialTransferMint mints tokens confidentially into the pending
// balance of destinationAccount.
//
// The three proofs are generated and inlined unless every proof account is set.
// A nil auditorElgamalPubkey stands for a mint configured without an auditor.
func ConfidentialTransferMint(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	destinationAccount solana.PublicKey,
	equalityProofAccount *solana.PublicKey,
	ciphertextValidityProofAccountWithCiphertext *ProofAccountWithCiphertext,
	rangeProofAccount *solana.PublicKey,
	mintAmount uint64,
	supplyElgamalKeypair *encryption.ElGamalKeypair,
	destinationElgamalPubkey encryption.ElGamalPubkey,
	auditorElgamalPubkey *encryption.ElGamalPubkey,
	aesKey zkencryption.AeKey,
	accountInfo SupplyAccountInfo,
) ([]solana.Instruction, error) {
	var proofs confidential.MintProofData
	if equalityProofAccount == nil ||
		ciphertextValidityProofAccountWithCiphertext == nil ||
		rangeProofAccount == nil {
		generated, err := accountInfo.GenerateSplitMintProofData(
			mintAmount, supplyElgamalKeypair, aesKey,
			destinationElgamalPubkey, auditorElgamalPubkey,
		)
		if err != nil {
			return nil, err
		}
		proofs = *generated
	}

	validity := proofs.CiphertextValidityProofDataWithCiphertext
	var ciphertextValidityProofAccount *solana.PublicKey
	if ciphertextValidityProofAccountWithCiphertext != nil {
		ciphertextValidityProofAccount = &ciphertextValidityProofAccountWithCiphertext.ContextStateAccount
		validity.CiphertextLo = ciphertextValidityProofAccountWithCiphertext.CiphertextLo
		validity.CiphertextHi = ciphertextValidityProofAccountWithCiphertext.CiphertextHi
	}

	newDecryptableSupply, err := accountInfo.NewDecryptableSupply(mintAmount, supplyElgamalKeypair, aesKey)
	if err != nil {
		return nil, err
	}
	return NewConfidentialMintBurnMintInstructions(
		destinationAccount, mint,
		validity.CiphertextLo, validity.CiphertextHi,
		authority, multisigSigners,
		zkprogram.ConfidentialTransferProofLocation(equalityProofAccount, 1, proofs.SupplyEqualityProofData),
		zkprogram.ConfidentialTransferProofLocation(ciphertextValidityProofAccount, 2, validity.ProofData),
		zkprogram.ConfidentialTransferProofLocation(rangeProofAccount, 3, proofs.RangeProofData),
		newDecryptableSupply,
	)
}

// ConfidentialTransferBurn burns tokens confidentially from the available
// balance of sourceAccount, adding the amount to the mint's pending burn.
//
// The three proofs are generated and inlined unless every proof account is set.
// A nil auditorElgamalPubkey stands for a mint configured without an auditor.
func ConfidentialTransferBurn(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	sourceAccount solana.PublicKey,
	equalityProofAccount *solana.PublicKey,
	ciphertextValidityProofAccountWithCiphertext *ProofAccountWithCiphertext,
	rangeProofAccount *solana.PublicKey,
	burnAmount uint64,
	sourceElgamalKeypair *encryption.ElGamalKeypair,
	supplyElgamalPubkey encryption.ElGamalPubkey,
	auditorElgamalPubkey *encryption.ElGamalPubkey,
	aesKey zkencryption.AeKey,
	accountInfo BurnAccountInfo,
) ([]solana.Instruction, error) {
	var proofs confidential.BurnProofData
	if equalityProofAccount == nil ||
		ciphertextValidityProofAccountWithCiphertext == nil ||
		rangeProofAccount == nil {
		generated, err := accountInfo.GenerateSplitBurnProofData(
			burnAmount, sourceElgamalKeypair, aesKey,
			supplyElgamalPubkey, auditorElgamalPubkey,
		)
		if err != nil {
			return nil, err
		}
		proofs = *generated
	}

	validity := proofs.CiphertextValidityProofDataWithCiphertext
	var ciphertextValidityProofAccount *solana.PublicKey
	if ciphertextValidityProofAccountWithCiphertext != nil {
		ciphertextValidityProofAccount = &ciphertextValidityProofAccountWithCiphertext.ContextStateAccount
		validity.CiphertextLo = ciphertextValidityProofAccountWithCiphertext.CiphertextLo
		validity.CiphertextHi = ciphertextValidityProofAccountWithCiphertext.CiphertextHi
	}

	newDecryptableBalance, err := accountInfo.NewDecryptableBalance(burnAmount, aesKey)
	if err != nil {
		return nil, err
	}
	return NewConfidentialMintBurnBurnInstructions(
		sourceAccount, mint, newDecryptableBalance,
		validity.CiphertextLo, validity.CiphertextHi,
		authority, multisigSigners,
		zkprogram.ConfidentialTransferProofLocation(equalityProofAccount, 1, proofs.EqualityProofData),
		zkprogram.ConfidentialTransferProofLocation(ciphertextValidityProofAccount, 2, validity.ProofData),
		zkprogram.ConfidentialTransferProofLocation(rangeProofAccount, 3, proofs.RangeProofData),
	)
}

// ConfidentialTransferApplyPendingBurn subtracts the pending burn of a mint from its confidential supply.
func ConfidentialTransferApplyPendingBurn(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) ([]solana.Instruction, error) {
	built, err := NewConfidentialMintBurnApplyPendingBurnInstruction(
		mint, authority, multisigSigners,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}
