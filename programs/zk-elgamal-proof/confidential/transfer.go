package confidential

import (
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

// Confidential transfer amounts are split into a 16-bit low and 32-bit high
// part, so a single transfer moves at most 2^48-1 tokens.
const (
	TransferAmountLoBitLength = 16
	TransferAmountHiBitLength = 32
	MaxTransferAmount         = 1<<(TransferAmountLoBitLength+TransferAmountHiBitLength) - 1
)

// CiphertextValidityProofWithAuditorCiphertext bundles a lo/hi ciphertext
// validity proof with the grouped ciphertexts it certifies.
type CiphertextValidityProofWithAuditorCiphertext struct {
	ProofData    *proofdata.BatchedGroupedCiphertext3HandlesValidityProofData
	CiphertextLo encryption.GroupedElGamalCiphertext3
	CiphertextHi encryption.GroupedElGamalCiphertext3
}

// TransferProofData is the proof data of confidential Transfer instruction.
type TransferProofData struct {
	// EqualityProofData proves the new balance ciphertext matches a commitment to the remaining amount.
	EqualityProofData *proofdata.CiphertextCommitmentEqualityProofData
	// CiphertextValidityProofDataWithCiphertext proves the lo/hi amount
	// ciphertexts are valid encryptions under the source, destination, and
	// auditor keys.
	CiphertextValidityProofDataWithCiphertext CiphertextValidityProofWithAuditorCiphertext
	// RangeProofData proves remaining and lo/hi amounts are in range.
	RangeProofData *proofdata.BatchedRangeProofU128Data
}

// TransferSplitProofData builds the three proofs a confidential Transfer
// requires: remaining-balance equality, lo/hi ciphertext validity under
// (source, destination, auditor), and the batched range proof.
//
// currentAvailableBalance is the source account's available balance
// ciphertext and currentDecryptableAvailableBalance its AE encryption under
// aeKey. A nil auditorPubkey stands in for a mint with no auditor.
func TransferSplitProofData(
	currentAvailableBalance encryption.ElGamalCiphertext,
	currentDecryptableAvailableBalance encryption.AeCiphertext,
	transferAmount uint64,
	sourceKeypair *encryption.ElGamalKeypair,
	aesKey encryption.AeKey,
	destinationPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
) (*TransferProofData, error) {
	currentBalanceAmount, err := aesKey.Decrypt(currentDecryptableAvailableBalance)
	if err != nil {
		return nil, err
	}
	if transferAmount > MaxTransferAmount {
		return nil, ErrIllegalAmountBitLength
	}
	if transferAmount > currentBalanceAmount {
		return nil, ErrNotEnoughFunds
	}
	transferAmountLo, transferAmountHi := splitAmount(transferAmount, TransferAmountLoBitLength)
	remainingBalance := currentBalanceAmount - transferAmount
	pubkeys := [3]encryption.ElGamalPubkey{sourceKeypair.Pubkey, destinationPubkey, orIdentity(auditorPubkey)}

	validityProof, transferAmountOpeningLo, transferAmountOpeningHi, err := encryptAndProveAmount(pubkeys, transferAmountLo, transferAmountHi)
	if err != nil {
		return nil, err
	}

	transferAmountCipherText, err := validityProof.combinedCiphertextForHandle(0, TransferAmountLoBitLength)
	if err != nil {
		return nil, err
	}

	remainingBalanceEqualityProof, remainingBalanceCommitment, remainingBalanceOpening, err := proveCiphertextDifference(
		sourceKeypair, currentAvailableBalance, transferAmountCipherText, remainingBalance)
	if err != nil {
		return nil, err
	}

	// Range proof over remainingBalance, transferAmountLo, transferAmountHi, and a zero pad totalling 128 bits.
	rangeProof, err := proveAmountRangeU128(
		remainingBalanceCommitment, remainingBalanceOpening, remainingBalance,
		transferAmountLo, transferAmountHi, transferAmountOpeningLo, transferAmountOpeningHi)
	if err != nil {
		return nil, err
	}

	return &TransferProofData{
		EqualityProofData:                         remainingBalanceEqualityProof,
		CiphertextValidityProofDataWithCiphertext: validityProof,
		RangeProofData:                            rangeProof,
	}, nil
}
