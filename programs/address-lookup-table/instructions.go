package addresslookuptable

import (
	"bytes"
	"encoding/binary"
	"fmt"

	spew "github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	text "github.com/gagliardetto/solana-go/text"
	treeout "github.com/gagliardetto/treeout"
)

var ProgramID solana.PublicKey = solana.AddressLookupTableProgramID

func SetProgramID(pubkey solana.PublicKey) {
	ProgramID = pubkey
	solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "AddressLookupTable"

func init() {
	solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const (
	// Create a new address lookup table.
	Instruction_CreateLookupTable uint32 = iota

	// Permanently freeze a lookup table, making it immutable.
	Instruction_FreezeLookupTable

	// Extend an address lookup table with new addresses.
	Instruction_ExtendLookupTable

	// Deactivate a lookup table, making it eligible for closing after a cooldown period.
	Instruction_DeactivateLookupTable

	// Close a deactivated lookup table and reclaim rent.
	Instruction_CloseLookupTable
)

// InstructionIDToName returns the name of the instruction given its ID.
func InstructionIDToName(id uint32) string {
	switch id {
	case Instruction_CreateLookupTable:
		return "CreateLookupTable"
	case Instruction_FreezeLookupTable:
		return "FreezeLookupTable"
	case Instruction_ExtendLookupTable:
		return "ExtendLookupTable"
	case Instruction_DeactivateLookupTable:
		return "DeactivateLookupTable"
	case Instruction_CloseLookupTable:
		return "CloseLookupTable"
	default:
		return ""
	}
}

type Instruction struct {
	bin.BaseVariant
}

func (inst *Instruction) EncodeToTree(parent treeout.Branches) {
	if enToTree, ok := inst.Impl.(text.EncodableToTree); ok {
		enToTree.EncodeToTree(parent)
	} else {
		parent.Child(spew.Sdump(inst))
	}
}

var InstructionImplDef = bin.NewVariantDefinition(
	bin.Uint32TypeIDEncoding,
	[]bin.VariantType{
		{"CreateLookupTable", (*CreateLookupTable)(nil)},
		{"FreezeLookupTable", (*FreezeLookupTable)(nil)},
		{"ExtendLookupTable", (*ExtendLookupTable)(nil)},
		{"DeactivateLookupTable", (*DeactivateLookupTable)(nil)},
		{"CloseLookupTable", (*CloseLookupTable)(nil)},
	},
)

func (inst *Instruction) ProgramID() solana.PublicKey {
	return ProgramID
}

func (inst *Instruction) Accounts() (out []*solana.AccountMeta) {
	return inst.Impl.(solana.AccountsGettable).GetAccounts()
}

func (inst *Instruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bin.NewBinEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

func (inst *Instruction) TextEncode(encoder *text.Encoder, option *text.Option) error {
	return encoder.Encode(inst.Impl, option)
}

func (inst *Instruction) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return inst.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionImplDef)
}

func (inst Instruction) MarshalWithEncoder(encoder *bin.Encoder) error {
	err := encoder.WriteUint32(inst.TypeID.Uint32(), binary.LittleEndian)
	if err != nil {
		return fmt.Errorf("unable to write variant type: %w", err)
	}
	return encoder.Encode(inst.Impl)
}

func registryDecodeInstruction(accounts []*solana.AccountMeta, data []byte) (any, error) {
	inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func DecodeInstruction(accounts []*solana.AccountMeta, data []byte) (*Instruction, error) {
	inst := new(Instruction)
	if err := bin.NewBinDecoder(data).Decode(inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction: %w", err)
	}
	if v, ok := inst.Impl.(solana.AccountsSettable); ok {
		err := v.SetAccounts(accounts)
		if err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}
