package confidential

import (
	"math"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

// MintProofData is the proof data a confidential Mint instruction carries.
type MintProofData struct {
	// SupplyEqualityProofData proves the new supply ciphertext matches a commitment
	// to the new supply amount.
	SupplyEqualityProofData *proofdata.CiphertextCommitmentEqualityProofData
	// CiphertextValidityProofDataWithCiphertext proves the lo/hi mint amount
	// ciphertexts are valid encryptions under the destination, supply, and
	// auditor keys, and carries the auditor's extracted lo/hi ciphertexts.
	CiphertextValidityProofDataWithCiphertext CiphertextValidityProofWithAuditorCiphertext
	// RangeProofData proves the new supply and lo/hi mint amounts are in range.
	RangeProofData *proofdata.BatchedRangeProofU128Data
}

// MintSplitProofData builds the three proofs a confidential Mint instruction requires.
//
// currentSupplyCiphertext is the mint's supply ciphertext and currentSupply its plaintext value.
// A nil auditorPubkey indicates no auditor
func MintSplitProofData(
	currentSupplyCiphertext encryption.ElGamalCiphertext,
	mintAmountPlaintext, currentSupplyPlaintext uint64,
	supplyKeypair *encryption.ElGamalKeypair,
	destinationPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
) (*MintProofData, error) {
	if mintAmountPlaintext > MaxAmount {
		return nil, ErrIllegalAmountBitLength
	}
	if mintAmountPlaintext > math.MaxUint64-currentSupplyPlaintext {
		return nil, ErrIllegalAmountBitLength
	}

	supplyChange, err := proveBalanceChange(supplyKeypair, currentSupplyCiphertext,
		[3]encryption.ElGamalPubkey{destinationPubkey, supplyKeypair.Pubkey, orIdentity(auditorPubkey)},
		1, mintAmountPlaintext, currentSupplyPlaintext+mintAmountPlaintext, proveCiphertextSum)
	if err != nil {
		return nil, err
	}

	// Range proof over new supply (64), lo (16), hi (32), and a zero pad
	// (16), totalling 128 bits.
	rangeProof, err := supplyChange.rangeProofU128()
	if err != nil {
		return nil, err
	}

	return &MintProofData{
		SupplyEqualityProofData:                   supplyChange.finalBalanceEqualityProof,
		CiphertextValidityProofDataWithCiphertext: supplyChange.changeAmountCipherTextValidityProof,
		RangeProofData:                            rangeProof,
	}, nil
}
