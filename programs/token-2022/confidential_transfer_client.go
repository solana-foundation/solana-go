package token2022

import (
	"errors"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/system"
	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/confidential"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/zkprogram"
)

// DefaultMaximumPendingBalanceCreditCounter is the default cap of pending balance credits.
const DefaultMaximumPendingBalanceCreditCounter uint64 = 65536

// ProofAccountWithCiphertext is a context state account holding a verified
// ciphertext validity proof, together with the transfer amount ciphertexts
// under the auditor key that proof was generated over.
type ProofAccountWithCiphertext struct {
	ContextStateAccount solana.PublicKey
	CiphertextLo        encryption.ElGamalCiphertext
	CiphertextHi        encryption.ElGamalCiphertext
}

// ConfidentialTransferCreateContextStateAccount creates a context state account and verifies proofData into it.
//
// rentLamports is the rent exemption balance for zkprogram.ContextStateSize
// bytes of the proof's context.
func ConfidentialTransferCreateContextStateAccount(
	contextStateAccount solana.PublicKey,
	contextStateAuthority solana.PublicKey,
	payer solana.PublicKey,
	proofData proofdata.ProofData,
	rentLamports uint64,
) ([]solana.Instruction, error) {
	if proofData == nil {
		return nil, errors.New("token2022: proof data not set")
	}
	verifyInstruction, err := zkprogram.ProofInstruction(proofData.ProofType()).EncodeVerifyProof(
		&zkprogram.ContextStateInfo{
			ContextStateAccount:   contextStateAccount,
			ContextStateAuthority: contextStateAuthority,
		},
		proofData,
	)
	if err != nil {
		return nil, err
	}
	createInstruction, err := system.NewCreateAccountInstruction(
		rentLamports,
		zkprogram.ContextStateSize(proofData.ContextData()),
		zkprogram.ProgramID,
		payer,
		contextStateAccount,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{createInstruction, verifyInstruction}, nil
}

// recordAccountDataOffset is where data starts in an account of the
// SPL record program: past the 1-byte version and 32-byte authority header.
const recordAccountDataOffset uint32 = 33

// ConfidentialTransferCreateContextStateAccountFromRecord creates a context
// state account and verifies proof data held in a record program account into it.
//
// rentLamports is the rent exemption balance for zkprogram.ContextStateSize
// bytes of the proof's context.
func ConfidentialTransferCreateContextStateAccountFromRecord(
	contextStateAccount solana.PublicKey,
	contextStateAuthority solana.PublicKey,
	payer solana.PublicKey,
	recordAccount solana.PublicKey,
	proofType proofdata.ProofType,
	rentLamports uint64,
) ([]solana.Instruction, error) {
	contextSize, err := zkprogram.ProofContextSize(proofType)
	if err != nil {
		return nil, err
	}
	verifyInstruction, err := zkprogram.ProofInstruction(proofType).EncodeVerifyProofFromAccount(
		&zkprogram.ContextStateInfo{
			ContextStateAccount:   contextStateAccount,
			ContextStateAuthority: contextStateAuthority,
		},
		recordAccount,
		recordAccountDataOffset,
	)
	if err != nil {
		return nil, err
	}
	createInstruction, err := system.NewCreateAccountInstruction(
		rentLamports,
		contextSize,
		zkprogram.ProgramID,
		payer,
		contextStateAccount,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{createInstruction, verifyInstruction}, nil
}

// ConfidentialTransferCloseContextStateAccount creates an instruction to close a proof context state
// account.
func ConfidentialTransferCloseContextStateAccount(
	contextStateAccount solana.PublicKey,
	lamportDestination solana.PublicKey,
	contextStateAuthority solana.PublicKey,
) solana.Instruction {
	return zkprogram.CloseContextStateInstruction(
		zkprogram.ContextStateInfo{
			ContextStateAccount:   contextStateAccount,
			ContextStateAuthority: contextStateAuthority,
		},
		lamportDestination,
	)
}

// ConfidentialTransferConfigureTokenAccount configures confidential transfers for a token account.
func ConfidentialTransferConfigureTokenAccount(
	account solana.PublicKey,
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	contextStateAccount *solana.PublicKey,
	maximumPendingBalanceCreditCounter *uint64,
	elgamalKeypair *encryption.ElGamalKeypair,
	aesKey zkencryption.AeKey,
) ([]solana.Instruction, error) {
	maximumCreditCounter := DefaultMaximumPendingBalanceCreditCounter
	if maximumPendingBalanceCreditCounter != nil {
		maximumCreditCounter = *maximumPendingBalanceCreditCounter
	}

	var proofData *proofdata.PubkeyValidityProofData
	if contextStateAccount == nil {
		var err error
		if proofData, err = proofdata.NewPubkeyValidityProofData(elgamalKeypair); err != nil {
			return nil, err
		}
	}

	decryptableZeroBalance, err := encryption.AeEncrypt(aesKey, 0)
	if err != nil {
		return nil, err
	}
	return NewConfidentialTransferConfigureAccountInstructions(
		account, mint, decryptableZeroBalance, maximumCreditCounter,
		authority, multisigSigners,
		zkprogram.ConfidentialTransferProofLocation(contextStateAccount, 1, proofData),
	)
}

// ConfidentialTransferConfigureTokenAccountWithRegistry configures
// confidential transfers for a token account from an ElGamal registry account:
// the account of the ElGamal registry program where the token account's owner
// has registered their ElGamal public key. The registered key was validated by
// the registry program, so unlike ConfidentialTransferConfigureTokenAccount
// this involves no proof and no owner signature; anybody can configure the
// account.
//
// A non-nil payer funds the reallocation growing the token account for the
// extension, and must sign.
func ConfidentialTransferConfigureTokenAccountWithRegistry(
	account solana.PublicKey,
	mint solana.PublicKey,
	elgamalRegistryAccount solana.PublicKey,
	payer *solana.PublicKey,
) ([]solana.Instruction, error) {
	built, err := NewConfidentialTransferConfigureAccountWithRegistryInstruction(
		account, mint, elgamalRegistryAccount, payer,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}

// ConfidentialTransferEmptyAccount prepares a token account with the
// confidential transfer extension for closing, proving that its available
// balance ciphertext encrypts zero.
func ConfidentialTransferEmptyAccount(
	account solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	contextStateAccount *solana.PublicKey,
	accountInfo EmptyAccountInfo,
	elgamalKeypair *encryption.ElGamalKeypair,
) ([]solana.Instruction, error) {
	var proofData *proofdata.ZeroCiphertextProofData
	if contextStateAccount == nil {
		var err error
		if proofData, err = accountInfo.GenerateProofData(elgamalKeypair); err != nil {
			return nil, err
		}
	}
	return NewConfidentialTransferEmptyAccountInstructions(
		account, authority, multisigSigners,
		zkprogram.ConfidentialTransferProofLocation(contextStateAccount, 1, proofData),
	)
}

// ConfidentialTransferDeposit deposits tokens from the non-confidential
// balance of a token account into its pending confidential balance.
func ConfidentialTransferDeposit(
	account solana.PublicKey,
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	amount uint64,
	decimals uint8,
) ([]solana.Instruction, error) {
	built, err := NewConfidentialTransferDepositInstruction(
		account, mint, amount, decimals, authority, multisigSigners,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}

// ConfidentialTransferApplyPendingBalance moves the pending confidential
// balance of a token account into its available balance.
func ConfidentialTransferApplyPendingBalance(
	account solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	accountInfo ApplyPendingBalanceAccountInfo,
	elgamalKeypair *encryption.ElGamalKeypair,
	aesKey zkencryption.AeKey,
) ([]solana.Instruction, error) {
	newDecryptableAvailableBalance, err := accountInfo.NewDecryptableAvailableBalance(elgamalKeypair, aesKey)
	if err != nil {
		return nil, err
	}
	built, err := NewConfidentialTransferApplyPendingBalanceInstruction(
		account, accountInfo.PendingBalanceCreditCounter(), newDecryptableAvailableBalance,
		authority, multisigSigners,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{built}, nil
}

// ConfidentialTransferWithdraw withdraws tokens from the available
// confidential balance of a token account into its non-confidential balance.
func ConfidentialTransferWithdraw(
	account solana.PublicKey,
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofAccount *solana.PublicKey,
	rangeProofAccount *solana.PublicKey,
	withdrawAmount uint64,
	decimals uint8,
	accountInfo WithdrawAccountInfo,
	elgamalKeypair *encryption.ElGamalKeypair,
	aesKey zkencryption.AeKey,
) ([]solana.Instruction, error) {
	var proofs confidential.WithdrawProofData
	if equalityProofAccount == nil || rangeProofAccount == nil {
		generated, err := accountInfo.GenerateProofData(withdrawAmount, elgamalKeypair, aesKey)
		if err != nil {
			return nil, err
		}
		proofs = *generated
	}

	newDecryptableAvailableBalance, err := accountInfo.NewDecryptableAvailableBalance(withdrawAmount, aesKey)
	if err != nil {
		return nil, err
	}
	return NewConfidentialTransferWithdrawInstructions(
		account, mint, withdrawAmount, decimals, newDecryptableAvailableBalance,
		authority, multisigSigners,
		zkprogram.ConfidentialTransferProofLocation(equalityProofAccount, 1, proofs.EqualityProofData),
		zkprogram.ConfidentialTransferProofLocation(rangeProofAccount, 2, proofs.RangeProofData),
	)
}

// ConfidentialTransferTransfer transfers tokens confidentially from the
// available balance of sourceAccount to the pending balance of destinationAccount.
func ConfidentialTransferTransfer(
	sourceAccount solana.PublicKey,
	destinationAccount solana.PublicKey,
	mint solana.PublicKey,
	sourceAuthority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofAccount *solana.PublicKey,
	ciphertextValidityProofAccountWithCiphertext *ProofAccountWithCiphertext,
	rangeProofAccount *solana.PublicKey,
	transferAmount uint64,
	accountInfo TransferAccountInfo,
	sourceElgamalKeypair *encryption.ElGamalKeypair,
	sourceAesKey zkencryption.AeKey,
	destinationElgamalPubkey encryption.ElGamalPubkey,
	auditorElgamalPubkey *encryption.ElGamalPubkey,
) ([]solana.Instruction, error) {
	var proofs confidential.TransferProofData
	if equalityProofAccount == nil ||
		ciphertextValidityProofAccountWithCiphertext == nil ||
		rangeProofAccount == nil {
		generated, err := accountInfo.GenerateSplitTransferProofData(
			transferAmount, sourceElgamalKeypair, sourceAesKey,
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

	newDecryptableAvailableBalance, err := accountInfo.NewDecryptableAvailableBalance(transferAmount, sourceAesKey)
	if err != nil {
		return nil, err
	}
	return NewConfidentialTransferTransferInstructions(
		sourceAccount, mint, destinationAccount, newDecryptableAvailableBalance,
		validity.CiphertextLo, validity.CiphertextHi,
		sourceAuthority, multisigSigners,
		zkprogram.ConfidentialTransferProofLocation(equalityProofAccount, 1, proofs.EqualityProofData),
		zkprogram.ConfidentialTransferProofLocation(ciphertextValidityProofAccount, 2, validity.ProofData),
		zkprogram.ConfidentialTransferProofLocation(rangeProofAccount, 3, proofs.RangeProofData),
	)
}

// ConfidentialTransferTransferWithFee transfers tokens confidentially on a
// mint extended for confidential transfer fees, withholding a fee of
// feeRateBasisPoints of the transfer amount, capped at maximumFee, in the
// destination account.
//
// The five proofs are generated and inlined unless every proof account is set.
// A nil auditorElgamalPubkey stands for a mint configured without an auditor.
func ConfidentialTransferTransferWithFee(
	sourceAccount solana.PublicKey,
	destinationAccount solana.PublicKey,
	mint solana.PublicKey,
	sourceAuthority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	equalityProofAccount *solana.PublicKey,
	ciphertextValidityProofAccountWithCiphertext *ProofAccountWithCiphertext,
	percentageWithCapProofAccount *solana.PublicKey,
	feeCiphertextValidityProofAccount *solana.PublicKey,
	rangeProofAccount *solana.PublicKey,
	transferAmount uint64,
	accountInfo TransferAccountInfo,
	sourceElgamalKeypair *encryption.ElGamalKeypair,
	sourceAesKey zkencryption.AeKey,
	destinationElgamalPubkey encryption.ElGamalPubkey,
	auditorElgamalPubkey *encryption.ElGamalPubkey,
	withdrawWithheldAuthorityElgamalPubkey encryption.ElGamalPubkey,
	feeRateBasisPoints uint16,
	maximumFee uint64,
) ([]solana.Instruction, error) {
	var proofs confidential.TransferWithFeeProofData
	if equalityProofAccount == nil ||
		ciphertextValidityProofAccountWithCiphertext == nil ||
		percentageWithCapProofAccount == nil ||
		feeCiphertextValidityProofAccount == nil ||
		rangeProofAccount == nil {
		generated, err := accountInfo.GenerateSplitTransferWithFeeProofData(
			transferAmount, sourceElgamalKeypair, sourceAesKey,
			destinationElgamalPubkey, auditorElgamalPubkey,
			withdrawWithheldAuthorityElgamalPubkey, feeRateBasisPoints, maximumFee,
		)
		if err != nil {
			return nil, err
		}
		proofs = *generated
	}

	validity := proofs.TransferAmountCiphertextValidityProofDataWithCiphertext
	var transferAmountCiphertextValidityProofAccount *solana.PublicKey
	if ciphertextValidityProofAccountWithCiphertext != nil {
		transferAmountCiphertextValidityProofAccount = &ciphertextValidityProofAccountWithCiphertext.ContextStateAccount
		validity.CiphertextLo = ciphertextValidityProofAccountWithCiphertext.CiphertextLo
		validity.CiphertextHi = ciphertextValidityProofAccountWithCiphertext.CiphertextHi
	}

	newDecryptableAvailableBalance, err := accountInfo.NewDecryptableAvailableBalance(transferAmount, sourceAesKey)
	if err != nil {
		return nil, err
	}
	return NewConfidentialTransferTransferWithFeeInstructions(
		sourceAccount, mint, destinationAccount, newDecryptableAvailableBalance,
		validity.CiphertextLo, validity.CiphertextHi,
		sourceAuthority, multisigSigners,
		zkprogram.ConfidentialTransferProofLocation(equalityProofAccount, 1, proofs.EqualityProofData),
		zkprogram.ConfidentialTransferProofLocation(transferAmountCiphertextValidityProofAccount, 2, validity.ProofData),
		zkprogram.ConfidentialTransferProofLocation(percentageWithCapProofAccount, 3, proofs.PercentageWithCapProofData),
		zkprogram.ConfidentialTransferProofLocation(feeCiphertextValidityProofAccount, 4, proofs.FeeCiphertextValidityProofData),
		zkprogram.ConfidentialTransferProofLocation(rangeProofAccount, 5, proofs.RangeProofData),
	)
}
