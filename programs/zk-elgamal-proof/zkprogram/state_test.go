package zkprogram

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

func TestProofContextStateRoundTrip(t *testing.T) {
	t.Parallel()
	authority := solana.NewWallet().PublicKey()

	for i := range NumProofTypes {
		typ := proofdata.ProofType(i + 1) // skip the unitialized proof type
		t.Run(typ.String(), func(t *testing.T) {
			t.Parallel()

			proof := fillProofData(t, typ)
			context := proof.ContextData()

			encoded, err := EncodeProofContextState(authority, typ, context)
			if err != nil {
				t.Fatalf("EncodeProofContextState: %v", err)
			}

			if got, want := len(encoded), int(ContextStateSize(context)); got != want {
				t.Fatalf("encoded length = %d, want %d", got, want)
			}
			if !bytes.Equal(encoded[:32], authority[:]) {
				t.Errorf("authority bytes = %x, want %x", encoded[:32], authority[:])
			}
			if encoded[32] != byte(typ) {
				t.Errorf("proof type byte = %d, want %d", encoded[32], byte(typ))
			}
			if got, want := encoded[ProofContextStateMetadataSize:], proof.Bytes()[:len(context.Bytes())]; !bytes.Equal(got, want) {
				t.Errorf("context bytes = %x, want %x", got, want)
			}

			var parsed ProofContextState
			if err := parsed.UnmarshalBinary(encoded); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}
			if parsed.ContextStateAuthority != authority {
				t.Errorf("authority = %s, want %s", parsed.ContextStateAuthority, authority)
			}
			if parsed.ProofType != typ {
				t.Errorf("proof type = %s, want %s", parsed.ProofType, typ)
			}
			if !bytes.Equal(parsed.Context, context.Bytes()) {
				t.Errorf("context = %x, want %x", parsed.Context, context.Bytes())
			}

			reparsed := newProofData(t, typ).ContextData()
			if err := reparsed.UnmarshalBinary(parsed.Context); err != nil {
				t.Fatalf("context UnmarshalBinary: %v", err)
			}
			if !bytes.Equal(reparsed.Bytes(), context.Bytes()) {
				t.Errorf("reparsed context = %x, want %x", reparsed.Bytes(), context.Bytes())
			}

			// The type-independent prefix parses without knowing the type.
			var meta ProofContextStateMetadata
			if err := meta.UnmarshalBinary(encoded); err != nil {
				t.Fatalf("metadata UnmarshalBinary: %v", err)
			}
			want := ProofContextStateMetadata{
				ContextStateAuthority: authority,
				ProofType:             typ,
			}
			if meta != want {
				t.Errorf("metadata = %+v, want %+v", meta, want)
			}
		})
	}
}

func TestProofContextStateUninitialized(t *testing.T) {
	t.Parallel()
	size := ContextStateSize(newProofData(t, proofdata.ProofTypeBatchedRangeProofU128).ContextData())

	var state ProofContextState
	if err := state.UnmarshalBinary(make([]byte, size)); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if state.ProofType != proofdata.ProofTypeUninitialized {
		t.Errorf("proof type = %s, want %s", state.ProofType, proofdata.ProofTypeUninitialized)
	}
	if state.ContextStateAuthority != (solana.PublicKey{}) {
		t.Errorf("authority = %s, want the zero key", state.ContextStateAuthority)
	}
	if want := make([]byte, size-ProofContextStateMetadataSize); !bytes.Equal(state.Context, want) {
		t.Errorf("context = %x, want %d zero bytes", state.Context, len(want))
	}
}

func TestProofContextStateRejects(t *testing.T) {
	t.Parallel()

	t.Run("short account data", func(t *testing.T) {
		t.Parallel()
		var state ProofContextState
		err := state.UnmarshalBinary(make([]byte, ProofContextStateMetadataSize-1))
		if err == nil || !strings.Contains(err.Error(), "want at least 33") {
			t.Fatalf("err = %v, want it to mention %q", err, "want at least 33")
		}
	})

	t.Run("unknown proof type", func(t *testing.T) {
		t.Parallel()
		data := make([]byte, 200)
		data[32] = 13
		var state ProofContextState
		if err := state.UnmarshalBinary(data); !errors.Is(err, proofdata.ErrInvalidProofType) {
			t.Fatalf("err = %v, want %v", err, proofdata.ErrInvalidProofType)
		}
	})

	t.Run("context of the wrong size", func(t *testing.T) {
		t.Parallel()
		data := make([]byte, ProofContextStateMetadataSize+16)
		data[32] = byte(proofdata.ProofTypePubkeyValidity)
		var state ProofContextState
		if err := state.UnmarshalBinary(data); err == nil || !strings.Contains(err.Error(), "want 32") {
			t.Fatalf("err = %v, want it to mention %q", err, "want 32")
		}
	})
}
