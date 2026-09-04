package token2022

import (
	"encoding"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// Encoded sizes of the field types appearing in ConfidentialTransfer
// instruction data.
const (
	accountKeySize        = solana.PublicKeyLength
	elGamalPubkeySize     = len(encryption.ElGamalPubkey{})
	elGamalCiphertextSize = len(encryption.ElGamalCiphertext{})
	aeCiphertextSize      = len(encryption.AeCiphertext{})
	u64Size               = 8
	u8Size                = 1
	boolSize              = 1
	proofOffsetSize       = 1
)

// newConfidentialTransferInstruction assembles a ConfidentialTransfer sub-instruction
func newConfidentialTransferInstruction(
	subInstruction uint8,
	data ag_binary.EncoderDecoder,
	accounts solana.AccountMetaSlice,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) *ConfidentialTransferExtension {
	authorityMeta := solana.Meta(authority)
	// Authority signs absent of multisig signers
	if len(multisigSigners) == 0 {
		authorityMeta.SIGNER()
	}
	ct_instruction := &ConfidentialTransferExtension{
		BaseVariant: ag_binary.BaseVariant{
			TypeID: ag_binary.TypeIDFromUint8(subInstruction),
			Impl:   data,
		},
		Accounts: append(accounts, authorityMeta),
		Signers:  make(solana.AccountMetaSlice, 0, len(multisigSigners)),
	}
	for _, signer := range multisigSigners {
		ct_instruction.Signers = append(ct_instruction.Signers, solana.Meta(signer).SIGNER())
	}
	return ct_instruction
}

// ctMarshalData and ctUnmarshalData adapt the fixed-size MarshalBinary and
// UnmarshalBinary implementations of the sub-instruction data structs to the
// ag_binary interfaces the variant machinery dispatches on.
func ctMarshalData(encoder *ag_binary.Encoder, d encoding.BinaryMarshaler) error {
	b, err := d.MarshalBinary()
	if err != nil {
		return err
	}
	return encoder.WriteBytes(b, false)
}

func ctUnmarshalData(decoder *ag_binary.Decoder, d encoding.BinaryUnmarshaler) error {
	b, err := decoder.ReadNBytes(decoder.Remaining())
	if err != nil {
		return err
	}
	return d.UnmarshalBinary(b)
}

// ctNoData is embedded by the data structs of sub-instructions that carry no
// data beyond the sub-instruction byte.
type ctNoData struct{}

func (ctNoData) MarshalWithEncoder(*ag_binary.Encoder) error { return nil }

func (*ctNoData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	if n := decoder.Remaining(); n != 0 {
		return fmt.Errorf("token2022: ConfidentialTransfer sub-instruction takes no data, got %d bytes", n)
	}
	return nil
}

// resolveProofLocation resolves a proof location to the account the consuming
// instruction carries for it and the offset to embed in the instruction data:
func resolveProofLocation[T proofdata.ProofData](
	location zkprogram.ProofLocation[T],
) (*solana.AccountMeta, int8, error) {
	if err := location.Validate(); err != nil {
		return nil, 0, err
	}
	// Embedded proofs return non-zero offset and the sysVarInstructionsAccount
	if location.IsInstructionOffset() {
		return solana.Meta(solana.SysVarInstructionsPubkey), location.InstructionOffset(), nil
	}
	// Context-state receipt return zero offset and the account holding the receipt
	return solana.Meta(location.ContextStateAccount()), 0, nil
}

// appendVerifyProofInstruction appends the VerifyProof instruction for a proof
// location in the instruction offset form.
//
// Callers must pass a slice whose element at index 0 is the consuming instruction,
// followed only by previously appended verify instructions.
func appendVerifyProofInstruction[T proofdata.ProofData](
	instructions []solana.Instruction,
	proofInstruction zkprogram.ProofInstruction,
	location zkprogram.ProofLocation[T],
) ([]solana.Instruction, error) {
	if !location.IsInstructionOffset() {
		return instructions, nil
	}
	expectedOffset := int8(len(instructions))
	if offset := location.InstructionOffset(); offset != expectedOffset {
		return nil, fmt.Errorf("token2022: proof instruction offset is %d, want %d", offset, expectedOffset)
	}
	verifyInstruction, err := proofInstruction.EncodeVerifyProof(nil, location.ProofData())
	if err != nil {
		return nil, err
	}
	return append(instructions, verifyInstruction), nil
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
