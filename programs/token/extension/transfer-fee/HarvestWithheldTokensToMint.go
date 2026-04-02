package transferfee

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// HarvestWithheldTokensToMint is a permissionless instruction to transfer all withheld tokens to the mint.
//
// Succeeds for frozen accounts.
// Accounts provided should include the TransferFeeAmount extension. If not, the account is skipped.
//
// Accounts expected by this instruction:
// - [writable] The mint.
// - [writable] The source accounts to harvest from.
type HarvestWithheldTokensToMint struct {
	// [0] = [WRITE] mint
	// ··········· The mint.
	//
	// [1...] = [WRITE] source_accounts
	// ··········· The source accounts to harvest from.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *HarvestWithheldTokensToMint) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	if len(accounts) < 1 {
		return errors.New("not enough accounts for HarvestWithheldTokensToMint")
	}
	obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	return nil
}

func (slice HarvestWithheldTokensToMint) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	return
}

func NewHarvestWithheldTokensToMintInstructionBuilder() *HarvestWithheldTokensToMint {
	nd := &HarvestWithheldTokensToMint{
		Accounts: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetMintAccount sets the "mint" account.
// The mint.
func (inst *HarvestWithheldTokensToMint) SetMintAccount(mint ag_solanago.PublicKey) *HarvestWithheldTokensToMint {
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint.
func (inst *HarvestWithheldTokensToMint) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// AddSourceAccount adds a source account to harvest from.
// The source accounts to harvest from.
func (inst *HarvestWithheldTokensToMint) AddSourceAccount(source ag_solanago.PublicKey) *HarvestWithheldTokensToMint {
	inst.Accounts = append(inst.Accounts, ag_solanago.Meta(source).WRITE())
	return inst
}

func (inst HarvestWithheldTokensToMint) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_HarvestWithheldTokensToMint),
	}}
}

func (inst HarvestWithheldTokensToMint) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *HarvestWithheldTokensToMint) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if len(inst.Accounts) < 2 {
			return errors.New("at least one source account must be provided")
		}
	}
	return nil
}

func (inst *HarvestWithheldTokensToMint) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("HarvestWithheldTokensToMint")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.Accounts[0]))

						sourceAccountsBranch := accountsBranch.Child(fmt.Sprintf("source_accounts[len=%v]", len(inst.Accounts)-1))
						for i, v := range inst.Accounts[1:] {
							if len(inst.Accounts)-1 > 9 && i < 10 {
								sourceAccountsBranch.Child(ag_format.Meta(fmt.Sprintf(" [%v]", i), v))
							} else {
								sourceAccountsBranch.Child(ag_format.Meta(fmt.Sprintf("[%v]", i), v))
							}
						}
					})
				})
		})
}

func (obj HarvestWithheldTokensToMint) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// This instruction has no parameters to serialize
	return nil
}

func (obj *HarvestWithheldTokensToMint) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// This instruction has no parameters to deserialize
	return nil
}
