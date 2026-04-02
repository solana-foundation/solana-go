package transferfee

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Transfers tokens from one account to another either directly or via a
// delegate.  If this account is associated with the native mint then equal
// amounts of SOL and Tokens will be transferred to the destination
// account.
//
// This instruction differs from Transfer in that the token mint and
// decimals value is checked by the caller.  This may be useful when
// creating transactions offline or within a hardware wallet.
type TransferCheckedWithFee struct {
	// The amount of tokens to transfer.
	Amount *uint64

	// Expected number of base 10 digits to the right of the decimal place.
	Decimals *uint8

	// Expected fee assessed on this transfer, calculated off-chain based on the transfer_fee_basis_points and maximum_fee of the mint.
	// May be 0 for a mint without a configured transfer fee.
	Fee *uint64

	// [0] = [WRITE] source
	// ··········· The source account.
	//
	// [1] = [] mint
	// ··········· The token mint.
	//
	// [2] = [WRITE] destination
	// ··········· The destination account.
	//
	// [3] = [] owner
	// ··········· The source account's owner/delegate.
	//
	// [4...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *TransferCheckedWithFee) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(4)
	return nil
}

func (slice TransferCheckedWithFee) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func NewTransferCheckedWithFeeInstructionBuilder() *TransferCheckedWithFee {
	nd := &TransferCheckedWithFee{
		Accounts: make(ag_solanago.AccountMetaSlice, 4),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

func (inst *TransferCheckedWithFee) SetAmount(amount uint64) *TransferCheckedWithFee {
	inst.Amount = &amount
	return inst
}

func (inst *TransferCheckedWithFee) SetDecimals(decimals uint8) *TransferCheckedWithFee {
	inst.Decimals = &decimals
	return inst
}

func (inst *TransferCheckedWithFee) SetFee(fee uint64) *TransferCheckedWithFee {
	inst.Fee = &fee
	return inst
}

// SetSourceAccount sets the "source" account.
// The source account.
func (inst *TransferCheckedWithFee) SetSourceAccount(source ag_solanago.PublicKey) *TransferCheckedWithFee {
	inst.Accounts[0] = ag_solanago.Meta(source).WRITE()
	return inst
}

// GetSourceAccount gets the "source" account.
// The source account.
func (inst *TransferCheckedWithFee) GetSourceAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetMintAccount sets the "mint" account.
// The token mint.
func (inst *TransferCheckedWithFee) SetMintAccount(mint ag_solanago.PublicKey) *TransferCheckedWithFee {
	inst.Accounts[1] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The token mint.
func (inst *TransferCheckedWithFee) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

func (inst *TransferCheckedWithFee) SetDestinationAccount(destination ag_solanago.PublicKey) *TransferCheckedWithFee {
	inst.Accounts[2] = ag_solanago.Meta(destination).WRITE()
	return inst
}

func (inst *TransferCheckedWithFee) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

func (inst *TransferCheckedWithFee) SetOwnerAccount(owner ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *TransferCheckedWithFee {
	inst.Accounts[3] = ag_solanago.Meta(owner)
	if len(multisigSigners) == 0 {
		inst.Accounts[3].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

func (inst *TransferCheckedWithFee) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[3]
}

func (inst TransferCheckedWithFee) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_TransferCheckedWithFee),
	}}
}

func (inst TransferCheckedWithFee) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *TransferCheckedWithFee) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Amount == nil {
			return errors.New("Amount parameter is not set")
		}
		if inst.Decimals == nil {
			return errors.New("Decimals parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Source is not set")
		}
		if inst.Accounts[1] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.Accounts[2] == nil {
			return errors.New("accounts.Destination is not set")
		}
		if inst.Accounts[3] == nil {
			return errors.New("accounts.Owner is not set")
		}
		if !inst.Accounts[3].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
		if len(inst.Signers) > MAX_SIGNERS {
			return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
		}
	}
	return nil
}

func (inst *TransferCheckedWithFee) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("TransferCheckedWithFee")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("  Amount", *inst.Amount))
						paramsBranch.Child(ag_format.Param("Decimals", *inst.Decimals))
						paramsBranch.Child(ag_format.Param("     Fee", *inst.Fee))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("     source", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("       mint", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("destination", inst.Accounts[2]))
						accountsBranch.Child(ag_format.Meta("      owner", inst.Accounts[3]))

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

func (obj TransferCheckedWithFee) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Amount` param:
	err = encoder.Encode(obj.Amount)
	if err != nil {
		return err
	}
	// Serialize `Decimals` param:
	err = encoder.Encode(obj.Decimals)
	if err != nil {
		return err
	}
	// Serialize `Fee` param:
	err = encoder.Encode(obj.Fee)
	if err != nil {
		return err
	}
	return nil
}
func (obj *TransferCheckedWithFee) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Amount`:
	err = decoder.Decode(&obj.Amount)
	if err != nil {
		return err
	}
	// Deserialize `Decimals`:
	err = decoder.Decode(&obj.Decimals)
	if err != nil {
		return err
	}
	// Deserialize `Fee`:
	err = decoder.Decode(&obj.Fee)
	if err != nil {
		return err
	}
	return nil
}
