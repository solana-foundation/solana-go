package zkprogram

import (
	"testing"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

func TestProofLocationValidate(t *testing.T) {
	t.Parallel()
	var zero ProofLocation[*proofdata.ZeroCiphertextProofData]
	if err := zero.Validate(); err == nil {
		t.Error("Validate accepted the zero value")
	}
	if err := ProofLocationInstructionOffset(1, (*proofdata.ZeroCiphertextProofData)(nil)).Validate(); err == nil {
		t.Error("Validate accepted a typed nil proof")
	}
	if err := ProofLocationInstructionOffset(1, &proofdata.ZeroCiphertextProofData{}).Validate(); err != nil {
		t.Errorf("Validate rejected an inline proof: %v", err)
	}
	if err := ProofLocationContextStateAccount[*proofdata.ZeroCiphertextProofData](solana.PublicKey{1}).Validate(); err != nil {
		t.Errorf("Validate rejected a context state account: %v", err)
	}
}

func TestEncodeVerifyProofRejectsTypedNil(t *testing.T) {
	t.Parallel()
	if _, err := VerifyZeroCiphertext.EncodeVerifyProof(nil, (*proofdata.ZeroCiphertextProofData)(nil)); err == nil {
		t.Error("EncodeVerifyProof accepted a typed nil proof")
	}
}
