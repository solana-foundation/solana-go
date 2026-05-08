package addresslookuptable

import (
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/binary"
	treeout "github.com/gagliardetto/treeout"
	solana "github.com/solana-foundation/solana-go/v2"
	format "github.com/solana-foundation/solana-go/v2/text/format"
)

// FreezeLookupTable permanently freezes a lookup table, making it immutable.
// The authority is removed and no further modifications are possible.
type FreezeLookupTable struct {
	// [0] = [WRITE] LookupTable
	// ··········· Address lookup table account to freeze
	//
	// [1] = [SIGNER] Authority
	// ··········· Current authority of the lookup table
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewFreezeLookupTableInstructionBuilder creates a new `FreezeLookupTable` instruction builder.
func NewFreezeLookupTableInstructionBuilder() *FreezeLookupTable {
	nd := &FreezeLookupTable{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
	return nd
}

func (inst *FreezeLookupTable) SetLookupTableAccount(lookupTable solana.PublicKey) *FreezeLookupTable {
	inst.AccountMetaSlice[0] = solana.Meta(lookupTable).WRITE()
	return inst
}

func (inst *FreezeLookupTable) SetAuthorityAccount(authority solana.PublicKey) *FreezeLookupTable {
	inst.AccountMetaSlice[1] = solana.Meta(authority).SIGNER()
	return inst
}

func (inst *FreezeLookupTable) GetLookupTableAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst *FreezeLookupTable) GetAuthorityAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst *FreezeLookupTable) SetAccounts(accounts []*solana.AccountMeta) error {
	inst.AccountMetaSlice = accounts
	if len(accounts) < 2 {
		return fmt.Errorf("insufficient accounts: FreezeLookupTable requires at least 2, got %d", len(accounts))
	}
	return nil
}

func (inst FreezeLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_FreezeLookupTable, binary.LittleEndian),
	}}
}

func (inst FreezeLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *FreezeLookupTable) Validate() error {
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *FreezeLookupTable) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("FreezeLookupTable")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch treeout.Branches) {})

					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("lookupTable", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta(" authority", inst.AccountMetaSlice[1]))
					})
				})
		})
}

func (inst FreezeLookupTable) MarshalWithEncoder(encoder *bin.Encoder) error {
	return nil
}

func (inst *FreezeLookupTable) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return nil
}

// NewFreezeLookupTableInstruction creates a new FreezeLookupTable instruction.
func NewFreezeLookupTableInstruction(
	lookupTable solana.PublicKey,
	authority solana.PublicKey,
) *FreezeLookupTable {
	return NewFreezeLookupTableInstructionBuilder().
		SetLookupTableAccount(lookupTable).
		SetAuthorityAccount(authority)
}
