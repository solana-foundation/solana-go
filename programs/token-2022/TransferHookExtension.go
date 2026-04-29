package token2022

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// TransferHook sub-instruction IDs.
const (
	TransferHook_Initialize uint8 = iota
	TransferHook_Update
)

// TransferHookExtension is the instruction wrapper for the TransferHook extension (ID 36).
type TransferHookExtension struct {
	SubInstruction uint8

	// The authority that can update the transfer hook.
	Authority *ag_solanago.PublicKey `bin:"-"`
	// The transfer hook program ID.
	HookProgramID *ag_solanago.PublicKey `bin:"-"`

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

func (obj *TransferHookExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	if obj.SubInstruction == TransferHook_Initialize {
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	} else {
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	}
	return nil
}

func (slice TransferHookExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst TransferHookExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_TransferHookExtension),
	}}
}

func (inst TransferHookExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *TransferHookExtension) Validate() error {
	if len(inst.Accounts) == 0 || inst.Accounts[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	if inst.SubInstruction == TransferHook_Update {
		if len(inst.Accounts) < 2 || inst.Accounts[1] == nil {
			return errors.New("accounts.Authority is not set")
		}
		if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	}
	return nil
}

func (inst *TransferHookExtension) EncodeToTree(parent ag_treeout.Branches) {
	name := "Initialize"
	if inst.SubInstruction == TransferHook_Update {
		name = "Update"
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("TransferHook." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						if inst.SubInstruction == TransferHook_Initialize {
							paramsBranch.Child(ag_format.Param("    Authority (OPT)", inst.Authority))
						}
						paramsBranch.Child(ag_format.Param("HookProgramID (OPT)", inst.HookProgramID))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("     mint", inst.Accounts[0]))
						if inst.SubInstruction == TransferHook_Update && len(inst.Accounts) > 1 {
							accountsBranch.Child(ag_format.Meta("authority", inst.Accounts[1]))
						}
					})
				})
		})
}

func writeOptionalPubkey(encoder *ag_binary.Encoder, pk *ag_solanago.PublicKey) error {
	if pk == nil {
		if err := encoder.WriteBool(false); err != nil {
			return err
		}
		empty := ag_solanago.PublicKey{}
		return encoder.WriteBytes(empty[:], false)
	}
	if err := encoder.WriteBool(true); err != nil {
		return err
	}
	return encoder.WriteBytes(pk[:], false)
}

func readOptionalPubkey(decoder *ag_binary.Decoder) (*ag_solanago.PublicKey, error) {
	ok, err := decoder.ReadBool()
	if err != nil {
		return nil, err
	}
	if ok {
		v, err := decoder.ReadNBytes(32)
		if err != nil {
			return nil, err
		}
		pk := ag_solanago.PublicKeyFromBytes(v)
		return &pk, nil
	}
	_, _ = decoder.ReadNBytes(32)
	return nil, nil
}

func (obj TransferHookExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.WriteUint8(obj.SubInstruction)
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case TransferHook_Initialize:
		if err = writeOptionalPubkey(encoder, obj.Authority); err != nil {
			return err
		}
		if err = writeOptionalPubkey(encoder, obj.HookProgramID); err != nil {
			return err
		}
	case TransferHook_Update:
		if err = writeOptionalPubkey(encoder, obj.HookProgramID); err != nil {
			return err
		}
	}
	return nil
}

func (obj *TransferHookExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case TransferHook_Initialize:
		obj.Authority, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
		obj.HookProgramID, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
	case TransferHook_Update:
		obj.HookProgramID, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewInitializeTransferHookInstruction creates an instruction to initialize the transfer hook extension.
func NewInitializeTransferHookInstruction(
	authority *ag_solanago.PublicKey,
	hookProgramId *ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
) *TransferHookExtension {
	inst := &TransferHookExtension{
		SubInstruction: TransferHook_Initialize,
		Authority:      authority,
		HookProgramID:  hookProgramId,
		Accounts:       make(ag_solanago.AccountMetaSlice, 1),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// NewUpdateTransferHookInstruction creates an instruction to update the transfer hook program ID.
func NewUpdateTransferHookInstruction(
	hookProgramId *ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *TransferHookExtension {
	inst := &TransferHookExtension{
		SubInstruction: TransferHook_Update,
		HookProgramID:  hookProgramId,
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
