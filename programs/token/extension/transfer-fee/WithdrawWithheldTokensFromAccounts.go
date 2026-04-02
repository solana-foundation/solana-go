package transferfee

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// WithdrawWithheldTokensFromAccounts transfers all withheld tokens to an account.
// Signed by the mint's withdraw withheld tokens authority.
//
// Single owner/delegate:
// - [] The token mint. Must include the TransferFeeConfig extension.
// - [writable] The fee receiver account. Must include the TransferFeeAmount extension and be associated with the provided mint.
// - [signer] The mint's withdraw_withheld_authority.
// - [writable] The source accounts to withdraw from.
//
// Multisignature owner/delegate:
// - [] The token mint.
// - [writable] The destination account.
// - [] The mint's multisig withdraw_withheld_authority.
// - [signer] M signer accounts.
// - [writable] The source accounts to withdraw from.
type WithdrawWithheldTokensFromAccounts struct {
	// Number of token accounts harvested
	NumTokenAccounts *uint8

	// [0] = [] mint
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
	//
	// [3+M...] = [WRITE] source_accounts
	// ··········· The source accounts to withdraw from.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *WithdrawWithheldTokensFromAccounts) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	return nil
}

func (slice WithdrawWithheldTokensFromAccounts) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func NewWithdrawWithheldTokensFromAccountsInstructionBuilder() *WithdrawWithheldTokensFromAccounts {
	nd := &WithdrawWithheldTokensFromAccounts{
		Accounts: make(ag_solanago.AccountMetaSlice, 3),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

func (inst *WithdrawWithheldTokensFromAccounts) SetNumTokenAccounts(numAccounts uint8) *WithdrawWithheldTokensFromAccounts {
	inst.NumTokenAccounts = &numAccounts
	return inst
}

// SetMintAccount sets the "mint" account.
// The token mint. Must include the TransferFeeConfig extension.
func (inst *WithdrawWithheldTokensFromAccounts) SetMintAccount(mint ag_solanago.PublicKey) *WithdrawWithheldTokensFromAccounts {
	inst.Accounts[0] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The token mint. Must include the TransferFeeConfig extension.
func (inst *WithdrawWithheldTokensFromAccounts) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetDestinationAccount sets the "destination" account.
// The fee receiver account (single owner) or destination account (multisig).
func (inst *WithdrawWithheldTokensFromAccounts) SetDestinationAccount(destination ag_solanago.PublicKey) *WithdrawWithheldTokensFromAccounts {
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	return inst
}

// GetDestinationAccount gets the "destination" account.
// The fee receiver account (single owner) or destination account (multisig).
func (inst *WithdrawWithheldTokensFromAccounts) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

// SetAuthorityAccount sets the "authority" account.
// The mint's withdraw_withheld_authority.
func (inst *WithdrawWithheldTokensFromAccounts) SetAuthorityAccount(authority ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *WithdrawWithheldTokensFromAccounts {
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
func (inst *WithdrawWithheldTokensFromAccounts) GetAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

// AddSourceAccount adds a source account to withdraw from.
// The source accounts to withdraw from.
func (inst *WithdrawWithheldTokensFromAccounts) AddSourceAccount(source ag_solanago.PublicKey) *WithdrawWithheldTokensFromAccounts {
	inst.Accounts = append(inst.Accounts, ag_solanago.Meta(source).WRITE())
	return inst
}

func (inst WithdrawWithheldTokensFromAccounts) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_WithdrawWithheldTokensFromAccounts),
	}}
}

func (inst WithdrawWithheldTokensFromAccounts) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *WithdrawWithheldTokensFromAccounts) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.NumTokenAccounts == nil {
			return errors.New("NumTokenAccounts parameter is not set")
		}
	}

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
		// Check that we have the expected number of source accounts
		expectedSourceAccounts := int(*inst.NumTokenAccounts)
		actualSourceAccounts := len(inst.Accounts) - 3
		if actualSourceAccounts != expectedSourceAccounts {
			return fmt.Errorf("expected %v source accounts, got %v", expectedSourceAccounts, actualSourceAccounts)
		}
	}
	return nil
}

func (inst *WithdrawWithheldTokensFromAccounts) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("WithdrawWithheldTokensFromAccounts")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("NumTokenAccounts", *inst.NumTokenAccounts))
					})

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

						sourceAccountsBranch := accountsBranch.Child(fmt.Sprintf("source_accounts[len=%v]", len(inst.Accounts)-3))
						for i, v := range inst.Accounts[3:] {
							if len(inst.Accounts)-3 > 9 && i < 10 {
								sourceAccountsBranch.Child(ag_format.Meta(fmt.Sprintf(" [%v]", i), v))
							} else {
								sourceAccountsBranch.Child(ag_format.Meta(fmt.Sprintf("[%v]", i), v))
							}
						}
					})
				})
		})
}

func (obj WithdrawWithheldTokensFromAccounts) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `NumTokenAccounts` param:
	err = encoder.Encode(obj.NumTokenAccounts)
	if err != nil {
		return err
	}
	return nil
}

func (obj *WithdrawWithheldTokensFromAccounts) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `NumTokenAccounts`:
	err = decoder.Decode(&obj.NumTokenAccounts)
	if err != nil {
		return err
	}
	return nil
}
