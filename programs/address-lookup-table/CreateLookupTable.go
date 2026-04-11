package addresslookuptable

import (
	"encoding/binary"
	"errors"
	"fmt"

	solana "github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	format "github.com/gagliardetto/solana-go/text/format"
	treeout "github.com/gagliardetto/treeout"
)

// CreateLookupTable creates a new address lookup table.
type CreateLookupTable struct {
	// Recent slot used to derive the lookup table address.
	RecentSlot *uint64

	// PDA bump seed.
	BumpSeed *uint8

	// [0] = [WRITE] LookupTable
	// ··········· Address lookup table account to create (derived PDA)
	//
	// [1] = [SIGNER] Authority
	// ··········· Authority that will control the lookup table
	//
	// [2] = [WRITE, SIGNER] Payer
	// ··········· Account that pays for the creation
	//
	// [3] = [] SystemProgram
	// ··········· System program
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewCreateLookupTableInstructionBuilder creates a new `CreateLookupTable` instruction builder.
func NewCreateLookupTableInstructionBuilder() *CreateLookupTable {
	nd := &CreateLookupTable{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
	return nd
}

func (inst *CreateLookupTable) SetRecentSlot(recentSlot uint64) *CreateLookupTable {
	inst.RecentSlot = &recentSlot
	return inst
}

func (inst *CreateLookupTable) SetBumpSeed(bumpSeed uint8) *CreateLookupTable {
	inst.BumpSeed = &bumpSeed
	return inst
}

func (inst *CreateLookupTable) SetLookupTableAccount(lookupTable solana.PublicKey) *CreateLookupTable {
	inst.AccountMetaSlice[0] = solana.Meta(lookupTable).WRITE()
	return inst
}

func (inst *CreateLookupTable) SetAuthorityAccount(authority solana.PublicKey) *CreateLookupTable {
	inst.AccountMetaSlice[1] = solana.Meta(authority).SIGNER()
	return inst
}

func (inst *CreateLookupTable) SetPayerAccount(payer solana.PublicKey) *CreateLookupTable {
	inst.AccountMetaSlice[2] = solana.Meta(payer).WRITE().SIGNER()
	return inst
}

func (inst *CreateLookupTable) SetSystemProgramAccount(systemProgram solana.PublicKey) *CreateLookupTable {
	inst.AccountMetaSlice[3] = solana.Meta(systemProgram)
	return inst
}

func (inst *CreateLookupTable) GetLookupTableAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst *CreateLookupTable) GetAuthorityAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst *CreateLookupTable) GetPayerAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}

func (inst *CreateLookupTable) GetSystemProgramAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[3]
}

func (inst *CreateLookupTable) SetAccounts(accounts []*solana.AccountMeta) error {
	inst.AccountMetaSlice = accounts
	if len(accounts) < 4 {
		return fmt.Errorf("insufficient accounts: CreateLookupTable requires at least 4, got %d", len(accounts))
	}
	return nil
}

func (inst CreateLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_CreateLookupTable, binary.LittleEndian),
	}}
}

func (inst CreateLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CreateLookupTable) Validate() error {
	if inst.RecentSlot == nil {
		return errors.New("RecentSlot parameter is not set")
	}
	if inst.BumpSeed == nil {
		return errors.New("BumpSeed parameter is not set")
	}
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *CreateLookupTable) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("CreateLookupTable")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("RecentSlot", *inst.RecentSlot))
						paramsBranch.Child(format.Param("  BumpSeed", *inst.BumpSeed))
					})

					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("  lookupTable", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("   authority", inst.AccountMetaSlice[1]))
						accountsBranch.Child(format.Meta("       payer", inst.AccountMetaSlice[2]))
						accountsBranch.Child(format.Meta("systemProgram", inst.AccountMetaSlice[3]))
					})
				})
		})
}

// DeriveLookupTableAddress derives the address and bump seed for an address lookup table
// from the given authority and recent slot.
func DeriveLookupTableAddress(authority solana.PublicKey, recentSlot uint64) (solana.PublicKey, uint8, error) {
	slotBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(slotBytes, recentSlot)
	return solana.FindProgramAddress(
		[][]byte{
			authority[:],
			slotBytes,
		},
		solana.AddressLookupTableProgramID,
	)
}

// NewCreateLookupTableInstruction creates a new CreateLookupTable instruction,
// deriving the lookup table address automatically.
// Returns the instruction builder and the derived lookup table address.
func NewCreateLookupTableInstruction(
	authority solana.PublicKey,
	payer solana.PublicKey,
	recentSlot uint64,
) (*CreateLookupTable, solana.PublicKey, error) {
	tableAddr, bumpSeed, err := DeriveLookupTableAddress(authority, recentSlot)
	if err != nil {
		return nil, solana.PublicKey{}, fmt.Errorf("failed to derive lookup table address: %w", err)
	}

	inst := NewCreateLookupTableInstructionBuilder().
		SetRecentSlot(recentSlot).
		SetBumpSeed(bumpSeed).
		SetLookupTableAccount(tableAddr).
		SetAuthorityAccount(authority).
		SetPayerAccount(payer).
		SetSystemProgramAccount(solana.SystemProgramID)

	return inst, tableAddr, nil
}
