package transferfee

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// WithdrawWithheldTokensFromMint transfers all withheld tokens in the mint to an account.
// Signed by the mint's withdraw withheld tokens authority.
//
// Single owner/delegate:
// - [writable] The token mint. Must include the TransferFeeConfig extension.
// - [writable] The fee receiver account. Must include the TransferFeeAmount extension associated with the provided mint.
// - [signer] The mint's withdraw_withheld_authority.
//
// Multisignature owner/delegate:
// - [writable] The token mint.
// - [writable] The destination account.
// - [] The mint's multisig withdraw_withheld_authority.
// - [signer] M signer accounts.
type WithdrawWithheldTokensFromMint struct {
	// [0] = [WRITE] mint
	// ··········· The token mint. Must include the TransferFeeConfig extension.
	//
	// [1] = [WRITE] destination
	// ··········· The fee receiver account (single owner) or destination account (multisig).
	//
	// [2] = [] authority
	// ··········· The mint's withdraw_withheld_authority.
	//
	// [3...] = [SIGNER] signers
	// ··········· M signer accounts (for multisig).
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *WithdrawWithheldTokensFromMint) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	return nil
}

func (slice WithdrawWithheldTokensFromMint) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func NewWithdrawWithheldTokensFromMintInstructionBuilder() *WithdrawWithheldTokensFromMint {
	nd := &WithdrawWithheldTokensFromMint{
		Accounts: make(ag_solanago.AccountMetaSlice, 3),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetMintAccount sets the "mint" account.
// The token mint. Must include the TransferFeeConfig extension.
func (inst *WithdrawWithheldTokensFromMint) SetMintAccount(mint ag_solanago.PublicKey) *WithdrawWithheldTokensFromMint {
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// GetMintAccount gets the "mint" account.
// The token mint. Must include the TransferFeeConfig extension.
func (inst *WithdrawWithheldTokensFromMint) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetDestinationAccount sets the "destination" account.
// The fee receiver account (single owner) or destination account (multisig).
func (inst *WithdrawWithheldTokensFromMint) SetDestinationAccount(destination ag_solanago.PublicKey) *WithdrawWithheldTokensFromMint {
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	return inst
}

// GetDestinationAccount gets the "destination" account.
// The fee receiver account (single owner) or destination account (multisig).
func (inst *WithdrawWithheldTokensFromMint) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

// SetAuthorityAccount sets the "authority" account.
// The mint's withdraw_withheld_authority.
func (inst *WithdrawWithheldTokensFromMint) SetAuthorityAccount(authority ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *WithdrawWithheldTokensFromMint {
	inst.Accounts[2] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// GetAuthorityAccount gets the "authority" account.
// The mint's withdraw_withheld_authority.
func (inst *WithdrawWithheldTokensFromMint) GetAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst WithdrawWithheldTokensFromMint) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_WithdrawWithheldTokensFromMint),
	}}
}

func (inst WithdrawWithheldTokensFromMint) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *WithdrawWithheldTokensFromMint) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.Accounts[1] == nil {
			return errors.New("accounts.Destination is not set")
		}
		if inst.Accounts[2] == nil {
			return errors.New("accounts.Authority is not set")
		}
		if !inst.Accounts[2].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
		if len(inst.Signers) > MAX_SIGNERS {
			return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
		}
	}
	return nil
}

func (inst *WithdrawWithheldTokensFromMint) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("WithdrawWithheldTokensFromMint")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("       mint", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("destination", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("   authority", inst.Accounts[2]))

						signersBranch := accountsBranch.Child(fmt.Sprintf("signers[len=%v]", len(inst.Signers)))
						for i, v := range inst.Signers {
							if len(inst.Signers) > 9 && i < 10 {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf(" [%v]", i), v))
							} else {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf("[%v]", i), v))
							}
						}
					})
				})
		})
}

func (obj WithdrawWithheldTokensFromMint) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// This instruction has no parameters to serialize
	return nil
}

func (obj *WithdrawWithheldTokensFromMint) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// This instruction has no parameters to deserialize
	return nil
}
