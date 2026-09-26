package token2022

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
)

// NewConfidentialMintBurnUpdateDecryptableSupplyInstruction updates the decryptable supply of a mint.
func NewConfidentialMintBurnUpdateDecryptableSupplyInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	newDecryptableSupply encryption.AeCiphertext,
) *ConfidentialMintBurnExtension {
	data := ConfidentialMintBurnUpdateDecryptableSupplyData{
		NewDecryptableSupply: newDecryptableSupply,
	}
	return newConfidentialMintBurnSubInstruction(
		ConfidentialMintBurn_UpdateDecryptableSupply,
		&data,
		solana.AccountMetaSlice{solana.Meta(mint).WRITE()},
		authority,
		multisigSigners,
	)
}

type ConfidentialMintBurnUpdateDecryptableSupplyData struct {
	NewDecryptableSupply encryption.AeCiphertext
}

const confidentialMintBurnUpdateDecryptableSupplyDataSize = aeCiphertextSize

func (d ConfidentialMintBurnUpdateDecryptableSupplyData) bytes() []byte {
	out := make([]byte, 0, confidentialMintBurnUpdateDecryptableSupplyDataSize)
	return append(out, d.NewDecryptableSupply[:]...)
}

func (d ConfidentialMintBurnUpdateDecryptableSupplyData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialMintBurnUpdateDecryptableSupplyData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialMintBurnUpdateDecryptableSupplyDataSize {
		return fmt.Errorf("token2022: ConfidentialMintBurn UpdateDecryptableSupply data is %d bytes, want %d", len(b), confidentialMintBurnUpdateDecryptableSupplyDataSize)
	}
	copy(d.NewDecryptableSupply[:], b)
	return nil
}
