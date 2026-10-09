package token2022

import (
	"math"

	"github.com/gagliardetto/solana-go/programs/token-2022/zkencryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/confidential"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
)

// --- Supply ---

// SupplyAccountInfo is the mint state a confidential Mint or RotateSupplyElGamalPubkey instruction is built from.
type SupplyAccountInfo struct {
	currentSupply       encryption.ElGamalCiphertext
	decryptableSupply   encryption.AeCiphertext
	supplyElGamalPubkey encryption.ElGamalPubkey
}

// NewSupplyAccountInfo extracts the state a Mint or RotateSupplyElGamalPubkey
// instruction needs from a confidential mint-burn mint state.
func NewSupplyAccountInfo(s *ConfidentialMintBurnState) SupplyAccountInfo {
	return SupplyAccountInfo{
		currentSupply:       encryption.ElGamalCiphertext(s.ConfidentialSupply),
		decryptableSupply:   encryption.AeCiphertext(s.DecryptableSupply),
		supplyElGamalPubkey: encryption.ElGamalPubkey(s.SupplyElGamalPubkey),
	}
}

// SupplyElGamalPubkey is the ElGamal pubkey the confidential supply is encrypted under.
func (i SupplyAccountInfo) SupplyElGamalPubkey() encryption.ElGamalPubkey {
	return i.supplyElGamalPubkey
}

// GenerateRotateSupplyElGamalPubkeyProof builds the proof a RotateSupplyElGamalPubkey instruction carries.
func (i SupplyAccountInfo) GenerateRotateSupplyElGamalPubkeyProof(
	currentSupplyElGamalKeypair *encryption.ElGamalKeypair,
	newSupplyElGamalPubkey encryption.ElGamalPubkey,
	aesKey zkencryption.AeKey,
) (*proofdata.CiphertextCiphertextEqualityProofData, error) {
	currentSupply, err := i.DecryptedCurrentSupply(aesKey, currentSupplyElGamalKeypair)
	if err != nil {
		return nil, err
	}
	return proveReencryption(currentSupplyElGamalKeypair, newSupplyElGamalPubkey, i.currentSupply, currentSupply)
}

// GenerateSplitMintProofData builds the three proofs a confidential Mint instruction requires.
// A nil auditorPubkey indicates no auditor.
func (i SupplyAccountInfo) GenerateSplitMintProofData(
	mintAmount uint64,
	supplyKeypair *encryption.ElGamalKeypair,
	aesKey zkencryption.AeKey,
	destinationPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
) (*confidential.MintProofData, error) {
	currentSupply, err := i.DecryptedCurrentSupply(aesKey, supplyKeypair)
	if err != nil {
		return nil, err
	}
	return confidential.MintSplitProofData(i.currentSupply, mintAmount, currentSupply,
		supplyKeypair, destinationPubkey, auditorPubkey)
}

// NewDecryptableSupply is the AE ciphertext a Mint instruction carries: the
// current supply plus the minted amount, encrypted with aesKey.
func (i SupplyAccountInfo) NewDecryptableSupply(
	mintAmount uint64, kp *encryption.ElGamalKeypair, aesKey zkencryption.AeKey,
) (encryption.AeCiphertext, error) {
	currentSupply, err := i.DecryptedCurrentSupply(aesKey, kp)
	if err != nil {
		return encryption.AeCiphertext{}, err
	}
	if mintAmount > math.MaxUint64-currentSupply {
		return encryption.AeCiphertext{}, confidential.ErrBalanceOverflow
	}
	return encryption.AeEncrypt(aesKey, currentSupply+mintAmount)
}

// DecryptedCurrentSupply recovers the current supply.
//
// The true supply balance is calculated by subtracting applied burns from cached supply balance.
func (i SupplyAccountInfo) DecryptedCurrentSupply(
	aesKey zkencryption.AeKey, kp *encryption.ElGamalKeypair,
) (uint64, error) {
	// Get cached plantext
	cachedSupply, err := encryption.AeDecrypt(aesKey, i.decryptableSupply)
	if err != nil {
		return 0, err
	}

	// re-encrypt cached supply to calculate the applied burns
	cachedSupplyCiphertext, err := kp.Pubkey.Encrypt(cachedSupply)
	if err != nil {
		return 0, err
	}
	appliedBurnsCiphertext, err := encryption.SubtractCiphertexts(cachedSupplyCiphertext, i.currentSupply)
	if err != nil {
		return 0, err
	}
	appliedBurns, err := kp.DecryptU32(appliedBurnsCiphertext)
	if err != nil {
		return 0, err
	}

	if appliedBurns > cachedSupply {
		return 0, confidential.ErrBalanceOverflow
	}
	return cachedSupply - appliedBurns, nil
}

// --- Burn ---

// BurnAccountInfo is the account state a confidential Burn instruction is built from.
type BurnAccountInfo struct {
	availableBalanceInfo
}

// NewBurnAccountInfo extracts the state a Burn instruction needs from a confidential transfer account.
func NewBurnAccountInfo(s *ConfidentialTransferAccountState) BurnAccountInfo {
	return BurnAccountInfo{availableBalanceInfo: newAvailableBalanceInfo(s)}
}

// GenerateSplitBurnProofData builds the three proofs a confidential Burn insstruction requires.
// A nil auditorPubkey indicates no auditor.
func (i BurnAccountInfo) GenerateSplitBurnProofData(
	burnAmount uint64,
	sourceKeypair *encryption.ElGamalKeypair,
	aesKey zkencryption.AeKey,
	supplyPubkey encryption.ElGamalPubkey,
	auditorPubkey *encryption.ElGamalPubkey,
) (*confidential.BurnProofData, error) {
	return confidential.BurnSplitProofData(i.availableBalance, i.decryptableAvailableBalance,
		burnAmount, sourceKeypair, aesKey, supplyPubkey, auditorPubkey)
}

// NewDecryptableBalance is the AE ciphertext a Burn instruction carries: the
// available balance less the burned amount, encrypted with aesKey.
func (i BurnAccountInfo) NewDecryptableBalance(
	burnAmount uint64, aesKey zkencryption.AeKey,
) (encryption.AeCiphertext, error) {
	return i.decryptableBalanceAfterDeduction(burnAmount, aesKey)
}
