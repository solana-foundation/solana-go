package confidential

import (
	"math"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

// Mint amounts are split into a 16-bit low and 32-bit high part.
const (
	MintAmountLoBits = 16
	MintAmountHiBits = 32
	MaxMintAmount    = 1<<(MintAmountLoBits+MintAmountHiBits) - 1 // 2^48 - 1
)

// MintProofData is the proof data of a confidential Mint instruction carries.
type MintProofData struct {
	// SupplyEqualityProofData proves the new supply ciphertext matches a commitment
	// to the new supply amount.
	SupplyEqualityProofData *proofdata.CiphertextCommitmentEqualityProofData
	// CiphertextValidityProofDataWithCiphertext proves the lo/hi mint amount
	// ciphertexts are valid encryptions under the destination, supply, and
	// auditor keys, and carries those ciphertexts.
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
	if mintAmountPlaintext > MaxMintAmount {
		return nil, ErrIllegalAmountBitLength
	}
	if mintAmountPlaintext > math.MaxUint64-currentSupplyPlaintext {
		return nil, ErrIllegalAmountBitLength
	}
	mintAmountLoPlaintext, mintAmountHiPlaintext := splitAmount(mintAmountPlaintext, MintAmountLoBits)
	newSupplyPlaintext := currentSupplyPlaintext + mintAmountPlaintext
	pubkeys := [3]encryption.ElGamalPubkey{destinationPubkey, supplyKeypair.Pubkey, orIdentity(auditorPubkey)}

	// Encrypt the mint amount split under all three keys and prove validity.
	mintAmountLoHiCiphertextValidityProof, openingLo, openingHi, err := encryptAndProveAmount(pubkeys, mintAmountLoPlaintext, mintAmountHiPlaintext)
	if err != nil {
		return nil, err
	}

	// New supply = current supply + (lo + 2^16 * hi)
	mintAmountCiphertext, err := mintAmountLoHiCiphertextValidityProof.combinedCiphertextForHandle(1, MintAmountLoBits)
	if err != nil {
		return nil, err
	}
	newSupplyCiphertextValidityProof, newSupplyCommitment, newSupplyOpening, err := proveCiphertextSum(
		supplyKeypair, currentSupplyCiphertext, mintAmountCiphertext, newSupplyPlaintext)
	if err != nil {
		return nil, err
	}

	// Range proof over new supply (64), lo (16), hi (32), and a zero pad
	// (16), totalling 128 bits.
	rangeProof, err := proveAmountRangeU128(
		newSupplyCommitment, newSupplyOpening, newSupplyPlaintext,
		mintAmountLoPlaintext, mintAmountHiPlaintext, openingLo, openingHi)
	if err != nil {
		return nil, err
	}

	return &MintProofData{
		SupplyEqualityProofData:                   newSupplyCiphertextValidityProof,
		CiphertextValidityProofDataWithCiphertext: mintAmountLoHiCiphertextValidityProof,
		RangeProofData:                            rangeProof,
	}, nil
}
