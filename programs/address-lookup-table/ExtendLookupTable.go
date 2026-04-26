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

// ExtendLookupTable extends an address lookup table with new addresses.
type ExtendLookupTable struct {
	// New addresses to add to the lookup table.
	Addresses []solana.PublicKey

	// [0] = [WRITE] LookupTable
	// ··········· Address lookup table account to extend
	//
	// [1] = [SIGNER] Authority
	// ··········· Current authority of the lookup table
	//
	// [2] = [WRITE, SIGNER] Payer (optional)
	// ··········· Account that pays for the extension (required when table needs more rent)
	//
	// [3] = [] SystemProgram (optional)
	// ··········· System program (required when payer is present)
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewExtendLookupTableInstructionBuilder creates a new `ExtendLookupTable` instruction builder.
func NewExtendLookupTableInstructionBuilder() *ExtendLookupTable {
	nd := &ExtendLookupTable{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2, 4),
	}
	return nd
}

func (inst *ExtendLookupTable) SetAddresses(addresses []solana.PublicKey) *ExtendLookupTable {
	inst.Addresses = addresses
	return inst
}

func (inst *ExtendLookupTable) SetLookupTableAccount(lookupTable solana.PublicKey) *ExtendLookupTable {
	inst.AccountMetaSlice[0] = solana.Meta(lookupTable).WRITE()
	return inst
}

func (inst *ExtendLookupTable) SetAuthorityAccount(authority solana.PublicKey) *ExtendLookupTable {
	inst.AccountMetaSlice[1] = solana.Meta(authority).SIGNER()
	return inst
}

func (inst *ExtendLookupTable) SetPayerAccount(payer solana.PublicKey) *ExtendLookupTable {
	if len(inst.AccountMetaSlice) < 4 {
		inst.AccountMetaSlice = append(inst.AccountMetaSlice, make(solana.AccountMetaSlice, 4-len(inst.AccountMetaSlice))...)
	}
	inst.AccountMetaSlice[2] = solana.Meta(payer).WRITE().SIGNER()
	inst.AccountMetaSlice[3] = solana.Meta(solana.SystemProgramID)
	return inst
}

func (inst *ExtendLookupTable) GetLookupTableAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst *ExtendLookupTable) GetAuthorityAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst *ExtendLookupTable) GetPayerAccount() *solana.AccountMeta {
	if len(inst.AccountMetaSlice) < 3 {
		return nil
	}
	return inst.AccountMetaSlice[2]
}

func (inst *ExtendLookupTable) GetSystemProgramAccount() *solana.AccountMeta {
	if len(inst.AccountMetaSlice) < 4 {
		return nil
	}
	return inst.AccountMetaSlice[3]
}

func (inst *ExtendLookupTable) SetAccounts(accounts []*solana.AccountMeta) error {
	inst.AccountMetaSlice = accounts
	if len(accounts) < 2 {
		return fmt.Errorf("insufficient accounts: ExtendLookupTable requires at least 2, got %d", len(accounts))
	}
	return nil
}

func (inst ExtendLookupTable) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_ExtendLookupTable, binary.LittleEndian),
	}}
}

func (inst ExtendLookupTable) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ExtendLookupTable) Validate() error {
	if len(inst.Addresses) == 0 {
		return errors.New("addresses parameter is empty")
	}
	// Check required accounts (first 2).
	for accIndex := 0; accIndex < 2; accIndex++ {
		if inst.AccountMetaSlice[accIndex] == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

func (inst *ExtendLookupTable) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("ExtendLookupTable")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child(fmt.Sprintf("Params[addresses=%d]", len(inst.Addresses))).
						ParentFunc(func(paramsBranch treeout.Branches) {
							for i, addr := range inst.Addresses {
								paramsBranch.Child(format.Param(fmt.Sprintf("Address[%d]", i), addr))
							}
						})

					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("lookupTable", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta(" authority", inst.AccountMetaSlice[1]))
						if inst.GetPayerAccount() != nil {
							accountsBranch.Child(format.Meta("     payer", inst.AccountMetaSlice[2]))
						}
						if inst.GetSystemProgramAccount() != nil {
							accountsBranch.Child(format.Meta("systemProgram", inst.AccountMetaSlice[3]))
						}
					})
				})
		})
}

func (inst ExtendLookupTable) MarshalWithEncoder(encoder *bin.Encoder) error {
	// Write address count as uint64 LE.
	if err := encoder.WriteUint64(uint64(len(inst.Addresses)), binary.LittleEndian); err != nil {
		return err
	}
	// Write each address.
	for _, addr := range inst.Addresses {
		if _, err := encoder.Write(addr[:]); err != nil {
			return err
		}
	}
	return nil
}

func (inst *ExtendLookupTable) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	count, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.Addresses = make([]solana.PublicKey, count)
	for i := uint64(0); i < count; i++ {
		_, err := decoder.Read(inst.Addresses[i][:])
		if err != nil {
			return err
		}
	}
	return nil
}

// NewExtendLookupTableInstruction creates a new ExtendLookupTable instruction.
func NewExtendLookupTableInstruction(
	lookupTable solana.PublicKey,
	authority solana.PublicKey,
	payer solana.PublicKey,
	addresses []solana.PublicKey,
) *ExtendLookupTable {
	return NewExtendLookupTableInstructionBuilder().
		SetAddresses(addresses).
		SetLookupTableAccount(lookupTable).
		SetAuthorityAccount(authority).
		SetPayerAccount(payer)
}
