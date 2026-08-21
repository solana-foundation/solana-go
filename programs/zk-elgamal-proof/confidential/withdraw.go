package confidential

import (
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

const BalanceBitLength = 64

// WithdrawProofData is the proof data a confidential Withdraw instruction.
type WithdrawProofData struct {
	// EqualityProofData proves the new balance ciphertext matches a
	// commitment to the remaining amount.
	EqualityProofData *proofdata.CiphertextCommitmentEqualityProofData
	// RangeProofData proves the remaining amount is a valid 64-bit value.
	RangeProofData *proofdata.BatchedRangeProofU64Data
}

// NewWithdrawProofData builds the proofs a confidential Withdraw requires:
// remaining-balance equality and the 64-bit range proof. The withdraw amount
// itself is public.
func NewWithdrawProofData(
	currentAvailableBalanceEncrypted encryption.ElGamalCiphertext,
	currentBalancePlaintext, withdrawalAmountPlainText uint64,
	keypair *encryption.ElGamalKeypair,
) (*WithdrawProofData, error) {
	if withdrawalAmountPlainText > currentBalancePlaintext {
		return nil, ErrNotEnoughFunds
	}
	remainingBalancePlaintext := currentBalancePlaintext - withdrawalAmountPlainText

	remainingBalanceCiphetext, err := currentAvailableBalanceEncrypted.SubtractAmount(withdrawalAmountPlainText)
	if err != nil {
		return nil, err
	}
	remainingBalanceCommitment, remainingBalanceOpening, err := encryption.NewPedersenCommitment(remainingBalancePlaintext)
	if err != nil {
		return nil, err
	}
	equalityProof, err := proofdata.NewCiphertextCommitmentEqualityProofData(
		keypair, remainingBalanceCiphetext, remainingBalanceCommitment, remainingBalanceOpening, remainingBalancePlaintext)
	if err != nil {
		return nil, err
	}
	rangeProof, err := proofdata.NewBatchedRangeProofU64Data(
		[]encryption.PedersenCommitment{remainingBalanceCommitment},
		[]uint64{remainingBalancePlaintext},
		[]uint8{BalanceBitLength},
		[]encryption.PedersenOpening{remainingBalanceOpening},
	)
	if err != nil {
		return nil, err
	}

	return &WithdrawProofData{
		EqualityProofData: equalityProof,
		RangeProofData:    rangeProof,
	}, nil
}
