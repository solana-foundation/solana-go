package confidential

import (
	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

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
// aesKey. A nil auditorPubkey stands in for a mint with no auditor.
func TransferSplitProofData(
	currentAvailableBalance encryption.ElGamalCiphertext,
	currentDecryptableAvailableBalance encryption.AeCiphertext,
	transferAmount uint64,
	sourceKeypair *encryption.ElGamalKeypair,
	aesKey zkencryption.AeKey,
	destinationPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
) (*TransferProofData, error) {
	currentBalanceAmount, err := encryption.AeDecrypt(aesKey, currentDecryptableAvailableBalance)
	if err != nil {
		return nil, err
	}
	if err := checkSpend(transferAmount, currentBalanceAmount); err != nil {
		return nil, err
	}

	change, err := proveBalanceChange(sourceKeypair, currentAvailableBalance,
		[3]encryption.ElGamalPubkey{sourceKeypair.Pubkey, destinationPubkey, orIdentity(auditorPubkey)},
		0, transferAmount, currentBalanceAmount-transferAmount, proveCiphertextDifference)
	if err != nil {
		return nil, err
	}

	// Range proof over remainingBalance, transferAmountLo, transferAmountHi, and a zero pad totalling 128 bits.
	rangeProof, err := change.rangeProofU128()
	if err != nil {
		return nil, err
	}

	return &TransferProofData{
		EqualityProofData:                         change.finalBalanceEqualityProof,
		CiphertextValidityProofDataWithCiphertext: change.changeAmountCipherTextValidityProof,
		RangeProofData:                            rangeProof,
	}, nil
}
