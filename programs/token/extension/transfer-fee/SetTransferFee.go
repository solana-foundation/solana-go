package transferfee

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// SetTransferFee sets the transfer fee. Only supported for mints that include the TransferFeeConfig extension.
//
// Single authority:
// - [writable] The mint.
// - [signer] The mint's fee account owner.
//
// Multisignature authority:
// - [writable] The mint.
// - [] The mint's multisignature fee account owner.
// - [signer] M signer accounts.
type SetTransferFee struct {
	// Amount of transfer collected as fees, expressed as basis points of the transfer amount
	TransferFeeBasisPoints *uint16

	// Maximum fee assessed on transfers
	MaximumFee *uint64

	// [0] = [WRITE] mint
	// ··········· The mint.
	//
	// [1] = [] authority
	// ··········· The mint's fee account owner.
	//
	// [2...] = [SIGNER] signers
	// ··········· M signer accounts (for multisig).
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *SetTransferFee) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	return nil
}

func (slice SetTransferFee) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func NewSetTransferFeeInstructionBuilder() *SetTransferFee {
	nd := &SetTransferFee{
		Accounts: make(ag_solanago.AccountMetaSlice, 2),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

func (inst *SetTransferFee) SetTransferFeeBasisPoints(basisPoints uint16) *SetTransferFee {
	inst.TransferFeeBasisPoints = &basisPoints
	return inst
}

func (inst *SetTransferFee) SetMaximumFee(maxFee uint64) *SetTransferFee {
	inst.MaximumFee = &maxFee
	return inst
}

// SetMintAccount sets the "mint" account.
// The mint.
func (inst *SetTransferFee) SetMintAccount(mint ag_solanago.PublicKey) *SetTransferFee {
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint.
func (inst *SetTransferFee) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetAuthorityAccount sets the "authority" account.
// The mint's fee account owner.
func (inst *SetTransferFee) SetAuthorityAccount(authority ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *SetTransferFee {
	inst.Accounts[1] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[1].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// GetAuthorityAccount gets the "authority" account.
// The mint's fee account owner.
func (inst *SetTransferFee) GetAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

func (inst SetTransferFee) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_SetTransferFee),
	}}
}

func (inst SetTransferFee) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *SetTransferFee) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.TransferFeeBasisPoints == nil {
			return errors.New("TransferFeeBasisPoints parameter is not set")
		}
		if inst.MaximumFee == nil {
			return errors.New("MaximumFee parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.Accounts[1] == nil {
			return errors.New("accounts.Authority is not set")
		}
		if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
		if len(inst.Signers) > MAX_SIGNERS {
			return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
		}
	}
	return nil
}

func (inst *SetTransferFee) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("SetTransferFee")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("TransferFeeBasisPoints", *inst.TransferFeeBasisPoints))
						paramsBranch.Child(ag_format.Param("MaximumFee", *inst.MaximumFee))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("      mint", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("  authority", inst.Accounts[1]))

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

func (obj SetTransferFee) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `TransferFeeBasisPoints` param:
	err = encoder.Encode(obj.TransferFeeBasisPoints)
	if err != nil {
		return err
	}
	// Serialize `MaximumFee` param:
	err = encoder.Encode(obj.MaximumFee)
	if err != nil {
		return err
	}
	return nil
}

func (obj *SetTransferFee) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `TransferFeeBasisPoints`:
	err = decoder.Decode(&obj.TransferFeeBasisPoints)
	if err != nil {
		return err
	}
	// Deserialize `MaximumFee`:
	err = decoder.Decode(&obj.MaximumFee)
	if err != nil {
		return err
	}
	return nil
}
