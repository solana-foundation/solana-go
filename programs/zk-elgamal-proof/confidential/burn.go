package confidential

import (
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

// Burn amounts are split into a 16-bit low part and 32-bit high part.
const (
	RemainingBalanceBitLength  = 64
	BurnAmountLoBitLength      = 16
	BurnAmountHiBitLength      = 32
	MaxBurnAmount              = 1<<(BurnAmountLoBitLength+BurnAmountHiBitLength) - 1
	RangeProofPaddingBitLength = 16
)

// BurnProofData is the proof data of a confidential Burn instruction.
type BurnProofData struct {
	// EqualityProofData proves the remaining balance ciphertext and commitment match.
	EqualityProofData *proofdata.CiphertextCommitmentEqualityProofData
	// CiphertextValidityProofDataWithCiphertext proves the lo/hi burn amount
	// ciphertexts are valid encryptions under the source, supply, and auditor
	// keys, and carries those ciphertexts.
	CiphertextValidityProofDataWithCiphertext CiphertextValidityProofWithAuditorCiphertext
	// RangeProofData proves remaining and lo/hi burn amounts fit in 128 bits.
	RangeProofData *proofdata.BatchedRangeProofU128Data
}

// BurnSplitProofData builds the three proofs a confidential Burn requires.
//
// currentAvailableBalanceCiphertext is the source account's available
// balance ciphertext and currentDecryptableAvailableBalance is its AE
// encryption under aeKey. A nil auditorPubkey indicates no auditor.
func BurnSplitProofData(
	currentAvailableBalanceEGCiphertext encryption.ElGamalCiphertext,
	currentDecryptableAvailableBalanceAESCiphertext encryption.AeCiphertext,
	burnAmountPlaintext uint64,
	sourceKeypair *encryption.ElGamalKeypair,
	aesKey encryption.AeKey,
	supplyPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
) (*BurnProofData, error) {
	currentBalancePlaintext, err := aesKey.Decrypt(currentDecryptableAvailableBalanceAESCiphertext)
	if err != nil {
		return nil, err
	}
	if burnAmountPlaintext > MaxBurnAmount {
		return nil, ErrIllegalAmountBitLength
	}
	if burnAmountPlaintext > currentBalancePlaintext {
		return nil, ErrNotEnoughFunds
	}
	burnAmountLoPlaintext, burnAmountHiPlaintext := splitAmount(burnAmountPlaintext, BurnAmountLoBitLength)
	remainingBalancePlaintext := currentBalancePlaintext - burnAmountPlaintext
	pubkeys := [3]encryption.ElGamalPubkey{sourceKeypair.Pubkey, supplyPubkey, orIdentity(auditorPubkey)}

	// Encrypt the burn amount split under all three keys and prove validity.
	burnAmountLoHiCiphertextValidityProof, openingLo, openingHi, err := encryptAndProveAmount(pubkeys, burnAmountLoPlaintext, burnAmountHiPlaintext)
	if err != nil {
		return nil, err
	}

	// New balance = current balance - (lo + 2^16 * hi) and
	// the equality proof that it matches a commitment to the remaining amount.
	burnAmountCiphertext, err := burnAmountLoHiCiphertextValidityProof.combinedCiphertextForHandle(0, BurnAmountLoBitLength)
	if err != nil {
		return nil, err
	}

	remainingBalanceEqualityProof, remainingBalanceCommitment, remainingBalanceOpening, err := proveCiphertextDifference(
		sourceKeypair, currentAvailableBalanceEGCiphertext, burnAmountCiphertext, remainingBalancePlaintext)
	if err != nil {
		return nil, err
	}

	// Range proof over remaining balance(64), lo (16), hi (32), and a zero pad (16) totalling 128 bits.
	rangeProof, err := proveAmountRangeU128(
		remainingBalanceCommitment, remainingBalanceOpening, remainingBalancePlaintext,
		burnAmountLoPlaintext, burnAmountHiPlaintext, openingLo, openingHi)
	if err != nil {
		return nil, err
	}

	return &BurnProofData{
		EqualityProofData:                         remainingBalanceEqualityProof,
		CiphertextValidityProofDataWithCiphertext: burnAmountLoHiCiphertextValidityProof,
		RangeProofData:                            rangeProof,
	}, nil
}
