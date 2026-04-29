package token2022

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// GroupMemberPointer sub-instruction IDs.
const (
	GroupMemberPointer_Initialize uint8 = iota
	GroupMemberPointer_Update
)

// GroupMemberPointerExtension is the instruction wrapper for the GroupMemberPointer extension (ID 41).
type GroupMemberPointerExtension struct {
	SubInstruction uint8

	Authority     *ag_solanago.PublicKey `bin:"-"`
	MemberAddress *ag_solanago.PublicKey `bin:"-"`

	// For Initialize:
	// [0] = [WRITE] mint
	//
	// For Update:
	// [0] = [WRITE] mint
	// [1] = [] authority
	// [2...] = [SIGNER] signers
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *GroupMemberPointerExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	if obj.SubInstruction == GroupMemberPointer_Initialize {
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	} else {
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	}
	return nil
}

func (slice GroupMemberPointerExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst GroupMemberPointerExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_GroupMemberPointerExtension),
	}}
}

func (inst GroupMemberPointerExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *GroupMemberPointerExtension) Validate() error {
	if len(inst.Accounts) == 0 || inst.Accounts[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	if inst.SubInstruction == GroupMemberPointer_Update {
		if len(inst.Accounts) < 2 || inst.Accounts[1] == nil {
			return errors.New("accounts.Authority is not set")
		}
		if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	}
	return nil
}

func (inst *GroupMemberPointerExtension) EncodeToTree(parent ag_treeout.Branches) {
	name := "Initialize"
	if inst.SubInstruction == GroupMemberPointer_Update {
		name = "Update"
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("GroupMemberPointer." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						if inst.SubInstruction == GroupMemberPointer_Initialize {
							paramsBranch.Child(ag_format.Param("    Authority (OPT)", inst.Authority))
						}
						paramsBranch.Child(ag_format.Param("MemberAddress (OPT)", inst.MemberAddress))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("     mint", inst.Accounts[0]))
						if inst.SubInstruction == GroupMemberPointer_Update && len(inst.Accounts) > 1 {
							accountsBranch.Child(ag_format.Meta("authority", inst.Accounts[1]))
						}
					})
				})
		})
}

func (obj GroupMemberPointerExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.WriteUint8(obj.SubInstruction)
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case GroupMemberPointer_Initialize:
		if err = writeOptionalPubkey(encoder, obj.Authority); err != nil {
			return err
		}
		if err = writeOptionalPubkey(encoder, obj.MemberAddress); err != nil {
			return err
		}
	case GroupMemberPointer_Update:
		if err = writeOptionalPubkey(encoder, obj.MemberAddress); err != nil {
			return err
		}
	}
	return nil
}

func (obj *GroupMemberPointerExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case GroupMemberPointer_Initialize:
		obj.Authority, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
		obj.MemberAddress, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
	case GroupMemberPointer_Update:
		obj.MemberAddress, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewInitializeGroupMemberPointerInstruction creates an instruction to initialize the group member pointer.
func NewInitializeGroupMemberPointerInstruction(
	authority *ag_solanago.PublicKey,
	memberAddress *ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
) *GroupMemberPointerExtension {
	inst := &GroupMemberPointerExtension{
		SubInstruction: GroupMemberPointer_Initialize,
		Authority:      authority,
		MemberAddress:  memberAddress,
		Accounts:       make(ag_solanago.AccountMetaSlice, 1),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// NewUpdateGroupMemberPointerInstruction creates an instruction to update the group member pointer.
func NewUpdateGroupMemberPointerInstruction(
	memberAddress *ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *GroupMemberPointerExtension {
	inst := &GroupMemberPointerExtension{
		SubInstruction: GroupMemberPointer_Update,
		MemberAddress:  memberAddress,
		Accounts:       make(ag_solanago.AccountMetaSlice, 2),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	inst.Accounts[1] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[1].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}
