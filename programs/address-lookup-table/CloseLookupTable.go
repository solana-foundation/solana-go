package addresslookuptable

import (
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/binary"
	treeout "github.com/gagliardetto/treeout"
	solana "github.com/solana-foundation/solana-go/v2"
	format "github.com/solana-foundation/solana-go/v2/text/format"
)

// CloseLookupTable closes a deactivated lookup table and reclaims
// rent to the recipient account.
type CloseLookupTable struct {
	// [0] = [WRITE] LookupTable
	// ··········· Address lookup table account to close
	//
	// [1] = [SIGNER] Authority
	// ··········· Current authority of the lookup table
	//
	// [2] = [WRITE] Recipient
	// ··········· Account to receive the reclaimed lamports
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewCloseLookupTableInstructionBuilder creates a new `CloseLookupTable` instruction builder.
func NewCloseLookupTableInstructionBuilder() *CloseLookupTable {
	nd := &CloseLookupTable{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
	return nd
}

func (inst *CloseLookupTable) SetLookupTableAccount(lookupTable solana.PublicKey) *CloseLookupTable {
	inst.AccountMetaSlice[0] = solana.Meta(lookupTable).WRITE()
	return inst
}

func (inst *CloseLookupTable) SetAuthorityAccount(authority solana.PublicKey) *CloseLookupTable {
	inst.AccountMetaSlice[1] = solana.Meta(authority).SIGNER()
	return inst
}

func (inst *CloseLookupTable) SetRecipientAccount(recipient solana.PublicKey) *CloseLookupTable {
	inst.AccountMetaSlice[2] = solana.Meta(recipient).WRITE()
	return inst
}

func (inst *CloseLookupTable) GetLookupTableAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst *CloseLookupTable) GetAuthorityAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst *CloseLookupTable) GetRecipientAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}

func (inst *CloseLookupTable) SetAccounts(accounts []*solana.AccountMeta) error {
	inst.AccountMetaSlice = accounts
	if len(accounts) < 3 {
		return fmt.Errorf("insufficient accounts: CloseLookupTable requires at least 3, got %d", len(accounts))
	}
	return nil
}

func (inst CloseLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_CloseLookupTable, binary.LittleEndian),
	}}
}

func (inst CloseLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CloseLookupTable) Validate() error {
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *CloseLookupTable) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("CloseLookupTable")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch treeout.Branches) {})

					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("lookupTable", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta(" authority", inst.AccountMetaSlice[1]))
						accountsBranch.Child(format.Meta(" recipient", inst.AccountMetaSlice[2]))
					})
				})
		})
}

func (inst CloseLookupTable) MarshalWithEncoder(encoder *bin.Encoder) error {
	return nil
}

func (inst *CloseLookupTable) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return nil
}

// NewCloseLookupTableInstruction creates a new CloseLookupTable instruction.
func NewCloseLookupTableInstruction(
	lookupTable solana.PublicKey,
	authority solana.PublicKey,
	recipient solana.PublicKey,
) *CloseLookupTable {
	return NewCloseLookupTableInstructionBuilder().
		SetLookupTableAccount(lookupTable).
		SetAuthorityAccount(authority).
		SetRecipientAccount(recipient)
}
