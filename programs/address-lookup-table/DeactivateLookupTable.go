package addresslookuptable

import (
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/binary"
	treeout "github.com/gagliardetto/treeout"
	solana "github.com/solana-foundation/solana-go/v2"
	format "github.com/solana-foundation/solana-go/v2/text/format"
)

// DeactivateLookupTable deactivates a lookup table, making it eligible
// for closing after a cooldown period (~512 slots).
type DeactivateLookupTable struct {
	// [0] = [WRITE] LookupTable
	// ··········· Address lookup table account to deactivate
	//
	// [1] = [SIGNER] Authority
	// ··········· Current authority of the lookup table
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewDeactivateLookupTableInstructionBuilder creates a new `DeactivateLookupTable` instruction builder.
func NewDeactivateLookupTableInstructionBuilder() *DeactivateLookupTable {
	nd := &DeactivateLookupTable{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
	return nd
}

func (inst *DeactivateLookupTable) SetLookupTableAccount(lookupTable solana.PublicKey) *DeactivateLookupTable {
	inst.AccountMetaSlice[0] = solana.Meta(lookupTable).WRITE()
	return inst
}

func (inst *DeactivateLookupTable) SetAuthorityAccount(authority solana.PublicKey) *DeactivateLookupTable {
	inst.AccountMetaSlice[1] = solana.Meta(authority).SIGNER()
	return inst
}

func (inst *DeactivateLookupTable) GetLookupTableAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst *DeactivateLookupTable) GetAuthorityAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst *DeactivateLookupTable) SetAccounts(accounts []*solana.AccountMeta) error {
	inst.AccountMetaSlice = accounts
	if len(accounts) < 2 {
		return fmt.Errorf("insufficient accounts: DeactivateLookupTable requires at least 2, got %d", len(accounts))
	}
	return nil
}

func (inst DeactivateLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_DeactivateLookupTable, binary.LittleEndian),
	}}
}

func (inst DeactivateLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *DeactivateLookupTable) Validate() error {
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *DeactivateLookupTable) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("DeactivateLookupTable")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch treeout.Branches) {})

					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("lookupTable", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta(" authority", inst.AccountMetaSlice[1]))
					})
				})
		})
}

func (inst DeactivateLookupTable) MarshalWithEncoder(encoder *bin.Encoder) error {
	return nil
}

func (inst *DeactivateLookupTable) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return nil
}

// NewDeactivateLookupTableInstruction creates a new DeactivateLookupTable instruction.
func NewDeactivateLookupTableInstruction(
	lookupTable solana.PublicKey,
	authority solana.PublicKey,
) *DeactivateLookupTable {
	return NewDeactivateLookupTableInstructionBuilder().
		SetLookupTableAccount(lookupTable).
		SetAuthorityAccount(authority)
}
