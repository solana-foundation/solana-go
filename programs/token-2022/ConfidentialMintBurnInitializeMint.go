package token2022

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
)

// NewConfidentialMintBurnInitializeMintInstruction initializes confidential mints and burns for a mint.
// It must appear in the same transaction as the InitializeMint instruction, otherwise another party can initialize the configuration.
func NewConfidentialMintBurnInitializeMintInstruction(
	mint solana.PublicKey,
	supplyElGamalPubkey encryption.ElGamalPubkey,
	decryptableSupply encryption.AeCiphertext,
) *ConfidentialMintBurnExtension {
	data := ConfidentialMintBurnInitializeMintData{
		SupplyElGamalPubkey: supplyElGamalPubkey,
		DecryptableSupply:   decryptableSupply,
	}
	return &ConfidentialMintBurnExtension{
		SubInstruction: ConfidentialMintBurn_InitializeMint,
		RawData:        data.bytes(),
		Accounts:       solana.AccountMetaSlice{solana.Meta(mint).WRITE()},
		Signers:        make(solana.AccountMetaSlice, 0),
	}
}

// ConfidentialMintBurnInitializeMintData is the instruction data for ConfidentialMintBurn_InitializeMint.
type ConfidentialMintBurnInitializeMintData struct {
	// SupplyElGamalPubkey is the ElGamal pubkey used to encrypt the confidential supply.
	SupplyElGamalPubkey encryption.ElGamalPubkey
	// DecryptableSupply is the initial zero supply encrypted with the supply AES key.
	DecryptableSupply encryption.AeCiphertext
}

const confidentialMintBurnInitializeMintDataSize = elGamalPubkeySize + aeCiphertextSize

func (d ConfidentialMintBurnInitializeMintData) bytes() []byte {
	out := make([]byte, 0, confidentialMintBurnInitializeMintDataSize)
	out = append(out, d.SupplyElGamalPubkey[:]...)
	out = append(out, d.DecryptableSupply[:]...)
	return out
}

func (d ConfidentialMintBurnInitializeMintData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialMintBurnInitializeMintData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialMintBurnInitializeMintDataSize {
		return fmt.Errorf("token2022: ConfidentialMintBurn InitializeMint data is %d bytes, want %d", len(b), confidentialMintBurnInitializeMintDataSize)
	}
	copy(d.SupplyElGamalPubkey[:], b[:elGamalPubkeySize])
	copy(d.DecryptableSupply[:], b[elGamalPubkeySize:])
	return nil
}
