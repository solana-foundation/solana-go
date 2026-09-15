package zkprogram

import (
	"strings"
	"testing"

	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

const NumProofTypes = 12

// newProofData returns a zero value of a proof data type.
func newProofData(t *testing.T, typ proofdata.ProofType) proofdata.ProofData {
	t.Helper()
	data := proofdata.NewProofData(typ)
	if data == nil {
		t.Fatalf("no proof data type for %s", typ)
	}
	return data
}

// fillProofData returns proof data of the given type whose every byte is
// distinct enough to catch a misordered or truncated serialization.
func fillProofData(t *testing.T, typ proofdata.ProofType) proofdata.ProofData {
	t.Helper()
	data := newProofData(t, typ)
	raw := make([]byte, len(data.Bytes()))
	for i := range raw {
		raw[i] = byte(i%251 + 1)
	}
	if err := data.UnmarshalBinary(raw); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	return data
}

// errorContains fails the test unless err is non-nil and mentions substr.
func errorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), substr) {
		t.Fatalf("err = %v, want it to mention %q", err, substr)
	}
}
