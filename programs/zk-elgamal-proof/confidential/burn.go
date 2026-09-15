package confidential

import (
	"github.com/solana-foundation/solana-go/v2/programs/token-2022/zkencryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/encryption"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

// BurnProofData is the proof data of a confidential Burn instruction.
type BurnProofData struct {
	// EqualityProofData proves the remaining balance ciphertext and commitment match.
	EqualityProofData *proofdata.CiphertextCommitmentEqualityProofData
	// CiphertextValidityProofDataWithCiphertext proves the lo/hi burn amount
	// ciphertexts are valid encryptions under the source, supply, and auditor
	// keys, and carries the auditor's extracted lo/hi ciphertexts.
	CiphertextValidityProofDataWithCiphertext CiphertextValidityProofWithAuditorCiphertext
	// RangeProofData proves remaining and lo/hi burn amounts fit in 128 bits.
	RangeProofData *proofdata.BatchedRangeProofU128Data
}

// BurnSplitProofData builds the three proofs a confidential Burn requires.
//
// currentAvailableBalanceCiphertext is the source account's available
// balance ciphertext and currentDecryptableAvailableBalance is its AE
// encryption under aesKey. A nil auditorPubkey indicates no auditor.
func BurnSplitProofData(
	currentAvailableBalanceEGCiphertext encryption.ElGamalCiphertext,
	currentDecryptableAvailableBalanceAESCiphertext encryption.AeCiphertext,
	burnAmountPlaintext uint64,
	sourceKeypair *encryption.ElGamalKeypair,
	aesKey zkencryption.AeKey,
	supplyPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
) (*BurnProofData, error) {
	currentBalancePlaintext, err := encryption.AeDecrypt(aesKey, currentDecryptableAvailableBalanceAESCiphertext)
	if err != nil {
		return nil, err
	}
	if err := checkSpend(burnAmountPlaintext, currentBalancePlaintext); err != nil {
		return nil, err
	}

	change, err := proveBalanceChange(sourceKeypair, currentAvailableBalanceEGCiphertext,
		[3]encryption.ElGamalPubkey{sourceKeypair.Pubkey, supplyPubkey, orIdentity(auditorPubkey)},
		0, burnAmountPlaintext, currentBalancePlaintext-burnAmountPlaintext, proveCiphertextDifference)
	if err != nil {
		return nil, err
	}

	rangeProof, err := change.rangeProofU128()
	if err != nil {
		return nil, err
	}

	return &BurnProofData{
		EqualityProofData:                         change.finalBalanceEqualityProof,
		CiphertextValidityProofDataWithCiphertext: change.changeAmountCipherTextValidityProof,
		RangeProofData:                            rangeProof,
	}, nil
}
