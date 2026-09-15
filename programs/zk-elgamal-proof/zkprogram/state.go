package zkprogram

import (
	"bytes"
	"fmt"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

// ProofContextState is the on-chain receipt of a verified proof.
type ProofContextState struct {
	// ContextStateAuthority is the unique account allowed to close the context state account.
	ContextStateAuthority solana.PublicKey
	// ProofType is the type of statement that is in the Context.
	ProofType proofdata.ProofType
	// Context is the pod serialization of the verified statement.
	Context []byte
}

// EncodeProofContextState encodes the account data a VerifyProof instruction
// writes for the given context.
func EncodeProofContextState(
	authority solana.PublicKey,
	proofType proofdata.ProofType,
	context proofdata.ProofContext,
) ([]byte, error) {
	meta, err := ProofContextStateMetadata{
		ContextStateAuthority: authority,
		ProofType:             proofType,
	}.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return append(meta, context.Bytes()...), nil
}

// UnmarshalBinary parses the data of a proof context state account.
func (s *ProofContextState) UnmarshalBinary(b []byte) error {
	var meta ProofContextStateMetadata
	if err := meta.UnmarshalBinary(b); err != nil {
		return err
	}
	if meta.ProofType != proofdata.ProofTypeUninitialized && !meta.ProofType.IsValid() {
		return fmt.Errorf("%w %d", proofdata.ErrInvalidProofType, meta.ProofType)
	}
	context := b[ProofContextStateMetadataSize:]
	// Check lengths of proof context for actual proofs
	// Unitialized proof-context lengths are unchecked.
	if meta.ProofType.IsValid() {
		want := len(proofdata.NewProofData(meta.ProofType).ContextData().Bytes())
		if len(context) != want {
			return fmt.Errorf("zk: %s proof context is %d bytes, want %d", meta.ProofType, len(context), want)
		}
	}
	s.ContextStateAuthority = meta.ContextStateAuthority
	s.ProofType = meta.ProofType
	s.Context = bytes.Clone(context)
	return nil
}

// ContextStateSize returns the exact number of bytes needed to represent a proof context state.
func ContextStateSize(context proofdata.ProofContext) uint64 {
	return uint64(ProofContextStateMetadataSize + len(context.Bytes()))
}

const ProofContextStateMetadataSize = solana.PublicKeyLength + 1

// ProofContextStateMetadata is the proof context state without the context data itself.
type ProofContextStateMetadata struct {
	// ContextStateAuthority is the unique account allowed to close the context state account.
	ContextStateAuthority solana.PublicKey
	// ProofType tags the proof context that follows.
	ProofType proofdata.ProofType
}

// MarshalBinary encodes the 33-byte prefix.
func (m ProofContextStateMetadata) MarshalBinary() ([]byte, error) {
	if m.ProofType > 0xFF {
		return nil, fmt.Errorf("%w %d", proofdata.ErrInvalidProofType, m.ProofType)
	}
	out := make([]byte, 0, ProofContextStateMetadataSize)
	out = append(out, m.ContextStateAuthority[:]...)
	return append(out, byte(m.ProofType)), nil
}

// UnmarshalBinary parses the 33-byte prefix, ignoring any proof context that follows it.
func (m *ProofContextStateMetadata) UnmarshalBinary(b []byte) error {
	if len(b) < ProofContextStateMetadataSize {
		return fmt.Errorf("zk: proof context state is %d bytes, want at least %d",
			len(b), ProofContextStateMetadataSize)
	}
	copy(m.ContextStateAuthority[:], b[:solana.PublicKeyLength])
	m.ProofType = proofdata.ProofType(b[solana.PublicKeyLength])
	return nil
}
