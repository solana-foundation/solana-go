package zkprogram

import (
	"bytes"
	"testing"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

// TestProofDataRoundTrip checks the raw-bytes contract.
func TestProofDataRoundTrip(t *testing.T) {
	t.Parallel()
	for i := range NumProofTypes {
		typ := proofdata.ProofType(i + 1) // skip the unitialized proof type
		t.Run(typ.String(), func(t *testing.T) {
			t.Parallel()
			proof := fillProofData(t, typ)
			proofInstruction := ProofInstruction(typ)

			inst, err := proofInstruction.EncodeVerifyProof(nil, proof)
			if err != nil {
				t.Fatalf("EncodeVerifyProof: %v", err)
			}
			data, err := inst.Data()
			if err != nil {
				t.Fatalf("Data: %v", err)
			}

			if _, err := DecodeInstruction(nil, data); err != nil {
				t.Errorf("DecodeInstruction: %v", err)
			}

			got, ok := InstructionType(data)
			if !ok {
				t.Fatal("InstructionType did not recognize the encoded data")
			}
			if got != proofInstruction {
				t.Errorf("instruction type = %v, want %v", got, proofInstruction)
			}

			parsed := newProofData(t, typ)
			if err := parsed.UnmarshalBinary(data[1:]); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}
			if !bytes.Equal(parsed.Bytes(), proof.Bytes()) {
				t.Errorf("parsed proof = %x, want %x", parsed.Bytes(), proof.Bytes())
			}

			// A truncated payload is caught where the caller names the type.
			errorContains(t, parsed.UnmarshalBinary(data[1:51]), "want")
		})
	}
}

func TestInstructionType(t *testing.T) {
	t.Parallel()
	if _, ok := InstructionType(nil); ok {
		t.Error("empty data: ok = true, want false")
	}
	if _, ok := InstructionType([]byte{13}); ok {
		t.Error("unknown discriminant: ok = true, want false")
	}
	typ, ok := InstructionType([]byte{0})
	if !ok {
		t.Fatal("discriminant 0: ok = false, want true")
	}
	if typ != CloseContextState {
		t.Errorf("type = %v, want %v", typ, CloseContextState)
	}
}

func TestRejectedEncodings(t *testing.T) {
	t.Parallel()
	proof := fillProofData(t, proofdata.ProofTypePubkeyValidity)

	t.Run("nil proof data", func(t *testing.T) {
		t.Parallel()
		_, err := VerifyPubkeyValidity.EncodeVerifyProof(nil, nil)
		errorContains(t, err, "proof data not set")
	})

	t.Run("mismatched proof data", func(t *testing.T) {
		t.Parallel()
		_, err := VerifyZeroCiphertext.EncodeVerifyProof(nil, proof)
		errorContains(t, err, "VerifyZeroCiphertext instruction cannot carry PubkeyValidity proof data")
	})

	t.Run("close cannot verify", func(t *testing.T) {
		t.Parallel()
		if _, err := CloseContextState.EncodeVerifyProof(nil, proof); err == nil {
			t.Fatal("EncodeVerifyProof on CloseContextState: err = nil, want error")
		}
		_, err := CloseContextState.EncodeVerifyProofFromAccount(nil, solana.PublicKey{}, 0)
		errorContains(t, err, "not a proof verification instruction")
	})

	t.Run("unknown discriminant", func(t *testing.T) {
		t.Parallel()
		_, err := ProofInstruction(13).EncodeVerifyProofFromAccount(nil, solana.PublicKey{}, 0)
		errorContains(t, err, "not a proof verification instruction")
	})
}

func TestDecodeInstruction(t *testing.T) {
	t.Parallel()
	proof := fillProofData(t, proofdata.ProofTypePubkeyValidity)
	inst, err := VerifyPubkeyValidity.EncodeVerifyProof(nil, proof)
	if err != nil {
		t.Fatalf("EncodeVerifyProof: %v", err)
	}
	data, err := inst.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}

	// The package registers itself so solana.DecodeInstruction resolves the
	// program.
	decoded, err := solana.DecodeInstruction(ProgramID, nil, data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	generic, ok := decoded.(*solana.GenericInstruction)
	if !ok {
		t.Fatalf("decoded instruction is %T, want *solana.GenericInstruction", decoded)
	}
	redecoded, err := generic.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if !bytes.Equal(redecoded, data) {
		t.Errorf("redecoded data = %x, want %x", redecoded, data)
	}

	_, err = DecodeInstruction(nil, []byte{13})
	errorContains(t, err, "not a proof program instruction")
	_, err = DecodeInstruction(nil, nil)
	errorContains(t, err, "not a proof program instruction")
	_, err = DecodeInstruction(nil, []byte{0, 1, 2, 3})
	errorContains(t, err, "CloseContextState takes no instruction data")
}
