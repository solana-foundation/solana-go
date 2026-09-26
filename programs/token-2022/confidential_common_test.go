package token2022

// Shared fixtures and assertion helpers for the tests of the confidential
// extensions (ConfidentialTransfer, ConfidentialTransferFee,
// ConfidentialMintBurn).

import (
	"bytes"
	"encoding"
	"reflect"
	"strings"
	"testing"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

func ctAddr(b byte) solana.PublicKey {
	var key solana.PublicKey
	for i := range key {
		key[i] = b
	}
	return key
}

// ctPattern fills n bytes the way the parity generator's pattern(seed, n) does.
func ctPattern(seed byte, n int) []byte {
	dst := make([]byte, n)
	for i := range dst {
		dst[i] = byte((i+int(seed))%251 + 1)
	}
	return dst
}

var (
	ctTokenAccount       = ctAddr(10)
	ctMint               = ctAddr(11)
	ctDestination        = ctAddr(12)
	ctAuthority          = ctAddr(13)
	ctMultisig           = []solana.PublicKey{ctAddr(14), ctAddr(15)}
	ctContextSingle      = ctAddr(20)
	ctContextEquality    = ctAddr(21)
	ctContextValidity    = ctAddr(22)
	ctContextFeeSigma    = ctAddr(23)
	ctContextFeeValidity = ctAddr(24)
	ctContextRange       = ctAddr(25)
	ctRegistry           = ctAddr(30)
	ctPayer              = ctAddr(31)
	ctRecord             = ctAddr(32)

	ctAuditorPubkey           = (*encryption.ElGamalPubkey)(ctPattern(1, 32))
	ctDecryptableBalance      = encryption.AeCiphertext(ctPattern(2, 36))
	ctCiphertextLo            = encryption.ElGamalCiphertext(ctPattern(3, 64))
	ctCiphertextHi            = encryption.ElGamalCiphertext(ctPattern(4, 64))
	cmbNewSupplyElGamalPubkey = encryption.ElGamalPubkey(ctPattern(5, 32))
	ctfSources                = []solana.PublicKey{ctAddr(40), ctAddr(41)}
)

const (
	ctAmount            = uint64(0x1122334455667788)
	ctDecimals          = uint8(9)
	ctMaxPendingCounter = uint64(65536)
)

// testConfidentialDataRoundTrip checks every instruction data value survives a
// marshal/unmarshal round trip and that off-by-one payload sizes are rejected.
func testConfidentialDataRoundTrip(t *testing.T, instructions []encoding.BinaryMarshaler) {
	t.Helper()
	for _, original := range instructions {
		t.Run(reflect.TypeOf(original).Name(), func(t *testing.T) {
			t.Parallel()
			raw, err := original.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}
			decoded := reflect.New(reflect.TypeOf(original))
			unmarshaler := decoded.Interface().(encoding.BinaryUnmarshaler)
			if err := unmarshaler.UnmarshalBinary(raw); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}
			if !reflect.DeepEqual(decoded.Elem().Interface(), original) {
				t.Errorf("round trip = %+v, want %+v", decoded.Elem().Interface(), original)
			}
			if len(raw) > 0 {
				if err := unmarshaler.UnmarshalBinary(raw[:len(raw)-1]); err == nil {
					t.Error("UnmarshalBinary accepted truncated data")
				}
			}
			if err := unmarshaler.UnmarshalBinary(append(raw, 0)); err == nil {
				t.Error("UnmarshalBinary accepted oversized data")
			}
		})
	}
}

// decodeBuiltConfidentialSubinstruction decodes a built instruction back through
// DecodeInstruction and returns the typed sub-instruction data it carries.
func decodeBuiltConfidentialSubinstruction[D confidentialSubInstructionData, E confidentialExtension[D]](
	t *testing.T, built *Instruction,
) D {
	t.Helper()
	data, err := built.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	decoded, err := DecodeInstruction(built.Accounts(), data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	wrapper, ok := decoded.Impl.(E)
	if !ok {
		var want E
		t.Fatalf("decoded to %T, want %T", decoded.Impl, want)
	}
	sub, err := decodeConfidentialSubInstruction[D](wrapper)
	if err != nil {
		t.Fatalf("DecodeSubInstructionData: %v", err)
	}
	return sub
}

// testConfidentialDecodeRejectsMalformed checks ext rejects a wrong-size
// payload, a payload on a dataless sub-instruction, and a sub-instruction ID
// the program does not define.
func testConfidentialDecodeRejectsMalformed[D confidentialSubInstructionData](
	t *testing.T,
	ext interface {
		confidentialExtension[D]
		UnmarshalWithDecoder(decoder *ag_binary.Decoder) error
	},
	sizedSubInstruction, datalessSubInstruction uint8,
) {
	t.Helper()
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"wrong size", []byte{sizedSubInstruction, 1, 2, 3}},
		{"data on dataless", []byte{datalessSubInstruction, 1}},
		{"unknown sub-instruction", []byte{200}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ext.UnmarshalWithDecoder(ag_binary.NewBinDecoder(tc.data)); err != nil {
				t.Fatalf("UnmarshalWithDecoder: %v", err)
			}
			if _, err := decodeConfidentialSubInstruction[D](ext); err == nil {
				t.Error("DecodeSubInstructionData accepted malformed sub-instruction data")
			}
		})
	}
}

// testConfidentialRejectsUnsetProofLocation checks build refuses the
// zero-value proof location and an offset location carrying typed nil proof
// data.
func testConfidentialRejectsUnsetProofLocation[T proofdata.ProofData](
	t *testing.T, build func(zkprogram.ProofLocation[T]) error,
) {
	t.Helper()
	var unset zkprogram.ProofLocation[T]
	if err := build(unset); err == nil {
		t.Error("builder accepted zero-value proof location")
	}
	var nilData T
	if err := build(zkprogram.ProofLocationInstructionOffset(1, nilData)); err == nil {
		t.Error("builder accepted typed nil proof data")
	}
}

// testConfidentialRawMatchesTyped checks the raw byte-level constructor
// encodes the same instruction data as the typed builder.
func testConfidentialRawMatchesTyped(t *testing.T, raw, typed interface {
	ValidateAndBuild() (*Instruction, error)
}) {
	t.Helper()
	rawBuilt, err := raw.ValidateAndBuild()
	if err != nil {
		t.Fatalf("raw ValidateAndBuild: %v", err)
	}
	typedBuilt, err := typed.ValidateAndBuild()
	if err != nil {
		t.Fatalf("typed ValidateAndBuild: %v", err)
	}
	rawData, err := rawBuilt.Data()
	if err != nil {
		t.Fatalf("raw Data: %v", err)
	}
	typedData, err := typedBuilt.Data()
	if err != nil {
		t.Fatalf("typed Data: %v", err)
	}
	if !bytes.Equal(rawData, typedData) {
		t.Errorf("raw constructor encoded %x, want %x", rawData, typedData)
	}
}

// wantProofOffsetError checks err reports an invalid proof instruction offset.
func wantProofOffsetError(t *testing.T, err error, label string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), "offset") {
		t.Errorf("%s err = %v, want proof instruction offset error", label, err)
	}
}
